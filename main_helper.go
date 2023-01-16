package zhelper

import (
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/labstack/echo/v4"
	"github.com/rs/xid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func Rs(c echo.Context, Ct map[string]interface{}) error {
	var Return Response
	empty := struct{}{}
	if KeyExists(Ct, "status") {
		Return.Status = Ct["status"].(int)
	} else {
		Return.Status = 200
	}
	if KeyExists(Ct, "message") {
		if Ct["message"] == "" {
			Return.Message = http.StatusText(Return.Status)
		} else {
			Return.Message = Ct["message"].(string)
		}
	} else {
		Return.Message = http.StatusText(Return.Status)
	}

	if KeyExists(Ct, "content") {
		Return.Content = Ct["content"]
	} else {
		Return.Content = empty
	}
	if KeyExists(Ct, "other") {
		Return.Others = Ct["other"]
	} else {
		Return.Others = empty
	}
	Return.Path = Substr(c.Request().RequestURI, 150)
	return c.JSON(Return.Status, Return)
}

func RsSuccess(c echo.Context) error {
	return Rs(c, H{
		"content": "success",
	})
}

func RsError(c echo.Context, code int, message interface{}) error {
	return Rs(c, H{
		"status":  code,
		"message": message,
	})
}

func KeyExists(decoded map[string]interface{}, key string) bool {
	val, ok := decoded[key]
	return ok && val != nil
}

// H is a shortcut for map[string]interface{}
type H map[string]interface{}

func ReadyBodyJson(c echo.Context, json map[string]interface{}) map[string]interface{} {
	if err := c.Bind(&json); err != nil {
		return nil
	}
	return json
}

func GetReq(Url string, token string) (*resty.Response, error) {
	client := resty.New()
	resp, err := client.R().EnableTrace().SetAuthToken(token).Get(Url)
	if err != nil {
		fmt.Println(err)
	}
	return resp, err
}

func JsonToMap(jsonStr string) map[string]interface{} {
	result := make(map[string]interface{})
	err := json.Unmarshal([]byte(jsonStr), &result)
	if err != nil {
		return nil
	}
	return result
}

func MarshalBinary(i interface{}) (data []byte) { //bytes to json string
	marshal, err := json.Marshal(i)
	if err != nil {
		println(err.Error())
	}
	return marshal
}

func RemoveField(obj interface{}, ignoreFields ...string) (interface{}, error) {
	toJson, err := json.Marshal(obj)
	if err != nil {
		return obj, err
	}
	if len(ignoreFields) == 0 {
		return obj, nil
	}
	toMap := map[string]interface{}{}
	json.Unmarshal(toJson, &toMap)
	for _, field := range ignoreFields {
		delete(toMap, field)
	}
	return toMap, nil
}

func IntString(result int) string {
	return strconv.Itoa(result)
}

func StringInt(result string) int {
	intVar, _ := strconv.Atoi(result)
	return intVar
}

func Substr(input string, limit int) string {
	if len([]rune(input)) >= limit {
		input = input[0:limit]
	}
	return input
}

func UniqueId() string {
	guid := xid.New()
	return guid.String()
}

func DateNow(typeFormat int) string {
	timeNow := time.Now()
	dateNow := time.Now().UTC()
	if typeFormat == 1 { //date only
		return dateNow.Format("2006-01-02")
	}
	if typeFormat == 2 { //time only
		return timeNow.Format("15:43:5")
	}
	if typeFormat == 3 { //datetime
		return dateNow.String()
	}
	return ""
}

func GormTime(timeParam time.Time) datatypes.Time {
	timeOnly := timeParam.Format("15:04:05")
	splitted := strings.Split(timeOnly, ":")
	return datatypes.NewTime(StringInt(splitted[0]), StringInt(splitted[1]), StringInt(splitted[2]), 0)
}

func DeletedAt() gorm.DeletedAt {
	return gorm.DeletedAt{
		Time:  time.Now(),
		Valid: true,
	}
}

func BlankString(stringText string) bool {
	if stringText == "" {
		return true
	}
	count := 0
	for _, v := range stringText {
		if v == ' ' {
			count++
		} else {
			break
		}
	}
	if count > 0 {
		return true
	}
	return false
}

// paginate helper
func GetParamPagination(c echo.Context) Pagination {
	// Get the query parameters from the request.
	query := c.Request().URL.Query()

	// Get the "limit", "page", "sort", and "asc" query parameters.
	// If the parameter is not present, the default value is returned.
	limit, _ := strconv.Atoi(query.Get("limit"))
	page, _ := strconv.Atoi(query.Get("page"))
	sort := query.Get("sort")
	asc, _ := strconv.Atoi(query.Get("asc"))

	// If the limit is not set or is set to 0, use a default value of 15.
	// If the limit is greater than 100, use a maximum value of 100.
	if limit == 0 {
		limit = 15
	}
	if limit > 100 {
		limit = 100
	}

	// If the page is not set or is set to 1, use a default value of 0.
	// Otherwise, decrement the page number by 1.
	if page <= 1 {
		page = 0
	} else {
		page--
	}

	// Set the sort order to "desc" by default.
	// If the "asc" query parameter is set to 1, set the sort order to "asc".
	ascFinal := "desc"
	if asc == 1 {
		ascFinal = "asc"
	}

	// If the "sort" query parameter is set, format it as `"field_name" order`.
	if sort != "" {
		sort = ToSnakeCase(sort)
		sort = `"` + sort + `" ` + ascFinal
	}

	// Return the pagination parameters as a Pagination struct.
	return Pagination{
		Limit:  limit,
		Page:   page,
		Sort:   sort,
		Search: c.QueryParam("search"),
		Field:  ToSnakeCase(c.QueryParam("field")),
	}
}

func Paginate(c echo.Context, qry *gorm.DB, total int64) (*gorm.DB, H) {
	pagination := GetParamPagination(c)
	offset := (pagination.Page) * pagination.Limit
	qryData := qry.Limit(pagination.Limit).Offset(offset)
	if pagination.Sort == "\"\" asc" {
		return qryData, PaginateInfo(pagination, total)
	}
	return qryData.Order(pagination.Sort), PaginateInfo(pagination, total)
}

func PaginateInfo(paging Pagination, totalData int64) H {
	// Calculate the total number of pages.
	totalPages := math.Ceil(float64(totalData) / float64(paging.Limit))

	// Calculate the next and previous page numbers.
	nextPage := paging.Page + 1
	if nextPage >= int(totalPages) {
		nextPage = 0
	}
	previousPage := paging.Page - 1
	if previousPage < 1 {
		previousPage = 0
	}

	// Increment the current page number.
	// If the current page is less than 1, set it to 1.
	paging.Page++
	if paging.Page < 1 {
		paging.Page = 1
	}

	// Return the pagination information as a map.
	return H{
		"nextPage":     nextPage,
		"previousPage": previousPage,
		"currentPage":  paging.Page,
		"totalPages":   totalPages,
		"totalData":    totalData,
	}
}

func ToSnakeCase(camel string) string {
	// Preallocate a slice of bytes with enough capacity to hold the final string.
	// This will avoid additional memory allocations and string copies when building the output string.
	buf := make([]byte, 0, len(camel)+5)

	// Iterate through the runes in the input string.
	for i := 0; i < len(camel); i++ {
		c := camel[i]
		if c >= 'A' && c <= 'Z' {
			// If the current rune is an uppercase letter, insert an underscore and convert it to lowercase.
			if len(buf) > 0 {
				buf = append(buf, '_')
			}
			buf = append(buf, c-'A'+'a')
		} else {
			// Otherwise, just append the current rune as is.
			buf = append(buf, c)
		}
	}
	return string(buf)
}

//end paginate helper
