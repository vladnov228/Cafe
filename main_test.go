package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCafeNegative(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []struct {
		request string
		status  int
		message string
	}{
		{"/cafe", http.StatusBadRequest, "unknown city"},
		{"/cafe?city=omsk", http.StatusBadRequest, "unknown city"},
		{"/cafe?city=tula&count=na", http.StatusBadRequest, "incorrect count"},
	}
	for _, v := range requests {
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", v.request, nil)
		handler.ServeHTTP(response, req)

		assert.Equal(t, v.status, response.Code)
		assert.Equal(t, v.message, strings.TrimSpace(response.Body.String()))
	}
}

func TestCafeWhenOk(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []string{
		"/cafe?count=2&city=moscow",
		"/cafe?city=tula",
		"/cafe?city=moscow&search=ложка",
	}
	for _, v := range requests {
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", v, nil)

		handler.ServeHTTP(response, req)

		assert.Equal(t, http.StatusOK, response.Code)
	}
}

func TestCafeCount(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []struct {
		count int
		want  int
	}{
		{count: 0, want: 0},
		{count: 1, want: 1},
		{count: 2, want: 2},
		{count: 100, want: len(cafeList["moscow"])},
	}

	for _, v := range requests {
		url := fmt.Sprintf("/cafe?city=moscow&count=%d", v.count)
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", url, nil)
		handler.ServeHTTP(response, req)
		require.Equal(t, http.StatusOK, response.Code)
		body := strings.TrimSpace(response.Body.String())

		var cafes []string
		if body != "" {
			cafes = strings.Split(body, ",")
		}
		assert.Len(t, cafes, v.want, "неверное количество кафе для count=%d", v.count)
	}
}

func TestCafeSearch(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []struct {
		search    string
		wantCount int
	}{
		{search: "фасоль", wantCount: 0},
		{search: "кофе", wantCount: 2},  // "Мир кофе", "Кофе и завтраки"
		{search: "вилка", wantCount: 1}, // "Ложка и вилка"
	}

	for _, v := range requests {
		url := fmt.Sprintf("/cafe?city=moscow&search=%s", v.search)
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", url, nil)
		handler.ServeHTTP(response, req)
		require.Equal(t, http.StatusOK, response.Code)
		body := strings.TrimSpace(response.Body.String())

		var cafes []string
		if body != "" {
			cafes = strings.Split(body, ",")
		}
		assert.Len(t, cafes, v.wantCount, "неверное количество найденных кафе для search=%s", v.search)
		for _, cafe := range cafes {
			assert.True(t,
				strings.Contains(strings.ToLower(cafe), strings.ToLower(v.search)),
				"кафе '%s' не содержит подстроку '%s'", cafe, v.search,
			)
		}
	}
}