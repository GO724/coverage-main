package main

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
)

// тут писать SearchServer

var DataFile string = "dataset.xml"

// Handler
func SearchServer(w http.ResponseWriter, r *http.Request) {

	// outputUsers := make([]User, 0, 100)
	// defer func(w http.ResponseWriter, u *[]User) {
	// 	// Write responce
	// 	json.NewEncoder(w).Encode(u)
	// 	fmt.Println(w)
	// }(w, &outputUsers)

	// Check authorization
	if r.Header.Get("AccessToken") != "valid_token" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Request parameters

	// Параметры offset и limit позволяют получать отсортированный список юзеров
	// пачками с индекса offset
	offset, err := strconv.Atoi(r.URL.Query().Get("offset")) // Pagination offset
	if err != nil || offset < 0 {
		offset = 0
	}

	// не более limit штук.
	limit, err := strconv.Atoi(r.URL.Query().Get("limit")) // Pagination limit
	if err != nil || limit < 1 {
		limit = 1
	}

	// Параметр order_field работает по полям Id, Age, Name
	// если пустой - то возвращаем по Name
	// если что-то другое - SearchServer ругается ошибкой.
	orderField := r.URL.Query().Get("order_field")
	switch orderField {
	case "Id", "Name", "Age":
	case "":
		orderField = "Name"
	default:
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(SearchErrorResponse{
			Error: ErrorBadOrderField,
		})
		return
	}

	// Параметр order_by задает направление сортировки (по полю переданному в order_field)
	// или ее отсутствие (OrderByAsIs = 0 	OrderByAsc  = 1	OrderByDesc = -1)
	orderBy, err := strconv.Atoi(r.URL.Query().Get("order_by"))
	if err != nil {
		orderBy = 0
	} else if orderBy < OrderByDesc {
		orderBy = OrderByDesc
	} else if orderBy > OrderByAsc {
		orderBy = OrderByAsc
	}

	query := r.URL.Query().Get("query")

	// load&parse : Данные для работы лежит в файле dataset.xml
	userStorage, err := loadDataset(DataFile)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Filter
	filteredUsers := filterUsers(userStorage, query)

	// Sort		orderAsc  = 1	orderDesc = -1
	if orderBy != 0 {
		err := sortUsers(filteredUsers, orderField, orderBy)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(SearchErrorResponse{
				Error: fmt.Sprintf("unexpected sortUsers error %+v", err),
			})
			return
		}
	}

	// Pagination
	total := len(filteredUsers)
	start := offset
	if start > total {
		start = total
	}
	end := start + limit
	if end > total {
		end = total
	}
	outputUsers := filteredUsers[start:end]
	// nextPage := end < total

	// Write responce
	json.NewEncoder(w).Encode(outputUsers)
}

func loadDataset(xmlPath string) ([]User, error) {

	// Данные для работы в файле dataset.xml
	xmlData, err := os.ReadFile(xmlPath)
	if err != nil {
		return []User{}, fmt.Errorf("getUser(): dataset.xml read error %w", err)
	}

	users, err := parseXMLData(xmlData)
	if err != nil {
		return []User{}, fmt.Errorf("getUser(): dataset.xml parse error %w", err)
	}
	return users, nil
}

// Name - это first_name + last_name из xml.

// <row>
//*	<id>0</id>
// 	<guid>1a6fa827-62f1-45f6-b579-aaead2b47169</guid>
// 	<isActive>false</isActive>
// 	<balance>$2,144.93</balance>
// 	<picture>http://placehold.it/32x32</picture>
//*	<age>22</age>
// 	<eyeColor>green</eyeColor>
//*	<first_name>Boyd</first_name>
//*	<last_name>Wolf</last_name>
//*	<gender>male</gender>
// 	<company>HOPELI</company>
// 	<email>boydwolf@hopeli.com</email>
// 	<phone>+1 (956) 593-2402</phone>
// 	<address>586 Winthrop Street, Edneyville, Mississippi, 9555</address>
//*	<about>Nulla cillum enim voluptate consequat laborum esse excepteur occaecat commodo nostrud excepteur ut cupidatat. Occaecat minim incididunt ut proident ad sint nostrud ad laborum sint pariatur. Ut nulla commodo dolore officia. Consequat anim eiusmod amet commodo eiusmod deserunt culpa. Ea sit dolore nostrud cillum proident nisi mollit est Lorem pariatur. Lorem aute officia deserunt dolor nisi aliqua consequat nulla nostrud ipsum irure id deserunt dolore. Minim reprehenderit nulla exercitation labore ipsum.</about>
// 	<registered>2017-02-05T06:23:27 -03:00</registered>
// 	<favoriteFruit>apple</favoriteFruit>
// </row>

// type User struct {
// 	ID     int
// 	Name   string  (first_name + last_name из xml)
// 	Age    int
// 	About  string
// 	Gender string
// }

// Temporary struct for parser XML
type xmlUser struct {
	XMLName   xml.Name `xml:"row"`
	ID        int      `xml:"id"`
	FirstName string   `xml:"first_name"`
	LastName  string   `xml:"last_name"`
	Age       int      `xml:"age"`
	About     string   `xml:"about"`
	Gender    string   `xml:"gender"`
}

// Temporary struct root
type xmlUsers struct {
	XMLName  xml.Name  `xml:"root"`
	XMLUsers []xmlUser `xml:"row"`
}

func parseXMLData(xmlData []byte) ([]User, error) {
	// Parse to tmp
	var xmlusers xmlUsers
	err := xml.Unmarshal(xmlData, &xmlusers)
	if err != nil {
		return nil, err
	}

	// Convert tmp_from_xml to []User
	result := make([]User, 0, len(xmlusers.XMLUsers))
	for _, u := range xmlusers.XMLUsers {
		result = append(result, User{
			ID:     u.ID,
			Name:   strings.TrimSpace(u.FirstName + " " + u.LastName),
			Age:    u.Age,
			About:  u.About,
			Gender: u.Gender,
		})
	}
	return result, nil
}

func filterUsers(users []User, query string) []User {
	if query == "" {
		return users
	}
	filtered := make([]User, 0)
	lowerQuery := strings.ToLower(query)
	for _, u := range users {
		if strings.Contains(strings.ToLower(u.Name), lowerQuery) ||
			strings.Contains(strings.ToLower(u.About), lowerQuery) {
			filtered = append(filtered, u)
		}
	}
	return filtered
}

func sortUsers(users []User, orderField string, orderBy int) error {
	// check valid orderField
	switch orderField {
	case "Id", "Name", "Age":
	default:
		return errors.New("invalid order field")
	}

	sort.Slice(users, func(i, j int) bool {
		switch orderField {
		case "Id":
			return (orderBy == 1 && users[i].ID < users[j].ID) || (orderBy == -1 && users[i].ID > users[j].ID)
		case "Name":
			return (orderBy == 1 && users[i].Name < users[j].Name) || (orderBy == -1 && users[i].Name > users[j].Name)
		case "Age":
			return (orderBy == 1 && users[i].Age < users[j].Age) || (orderBy == -1 && users[i].Age > users[j].Age)
		}
		return false
	})
	return nil
}
