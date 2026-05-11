package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCreateFolderWithoutAuth(t *testing.T) {
	data := bytes.NewBufferString(`{"name":"anime"}`)
	req, err := http.NewRequest(http.MethodPost, address+"/api/folders", data)
	require.NoError(t, err, "cant make request")

	resp, err := client.Do(req)
	require.NoError(t, err, "could not send login command")
	defer resp.Body.Close()

	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestCreateFolders(t *testing.T) {
	registerUser(t, "dima", "marvel")
	registerUser(t, "arseniy", "qwerty")
	tokenDima := loginNoAdmin(t, "dima", "marvel")
	tokenArseniy := loginNoAdmin(t, "arseniy", "qwerty")

	testscase := []struct {
		name     string
		data     string
		wantCode int
		dima     bool
		arseniy  bool
	}{
		{
			name:     "OK_1",
			data:     `{"name":"anime"}`,
			wantCode: http.StatusOK,
			dima:     true,
			arseniy:  true,
		},

		{
			name:     "OK_2",
			data:     `{"name":"manga"}`,
			wantCode: http.StatusOK,
			dima:     true,
		},

		{
			name:     "OK_3",
			data:     `{"name":"genshin"}`,
			wantCode: http.StatusOK,
			dima:     true,
			arseniy:  true,
		},

		{
			name:     "OK_4",
			data:     `{"name":"magbitva"}`,
			wantCode: http.StatusOK,
			arseniy:  true,
		},

		{
			name:     "SameFolder",
			data:     `{"name":"anime"}`,
			wantCode: http.StatusConflict,
			dima:     true,
			arseniy:  true,
		},

		{
			name:     "Empty",
			data:     `{"name":""}`,
			wantCode: http.StatusBadRequest,
			dima:     true,
			arseniy:  true,
		},
	}

	for _, tc := range testscase {
		t.Run(tc.name, func(t *testing.T) {
			if tc.dima {
				req1, err := http.NewRequest(http.MethodPost, address+"/api/folders", bytes.NewBufferString(tc.data))
				require.NoError(t, err, "cant create req")
				req1.Header.Add("Authorization", "Token "+tokenDima)
				resp, err := client.Do(req1)
				require.NoError(t, err, "cant do request")
				defer resp.Body.Close()
				require.Equal(t, tc.wantCode, resp.StatusCode)
			}

			if tc.arseniy {
				req2, err := http.NewRequest(http.MethodPost, address+"/api/folders", bytes.NewBufferString(tc.data))
				require.NoError(t, err, "cant create req")
				req2.Header.Add("Authorization", "Token "+tokenArseniy)
				resp, err := client.Do(req2)
				require.NoError(t, err, "cant do request")
				defer resp.Body.Close()
				require.Equal(t, tc.wantCode, resp.StatusCode)
			}
		})
	}

}

type Folder struct {
	FolderId int    `json:"folder_id"`
	Name     string `json:"name"`
}

type ListFolders struct {
	Folders []Folder `json:"folders"`
}

func TestListFolders(t *testing.T) {
	tokenDima := loginNoAdmin(t, "dima", "marvel")

	req1, err := http.NewRequest(http.MethodGet, address+"/api/folders", nil)
	require.NoError(t, err, "cant create req")
	req1.Header.Add("Authorization", "Token "+tokenDima)
	resp, err := client.Do(req1)
	require.NoError(t, err, "cant do request")
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var folders ListFolders

	err = json.NewDecoder(resp.Body).Decode(&folders)
	require.NoError(t, err)

	require.Equal(t, 3, len(folders.Folders))

	arr := []string{"anime", "manga", "genshin"}
	for i := range 3 {
		require.Contains(t, arr, folders.Folders[i].Name)
	}

	tokenArseniy := loginNoAdmin(t, "arseniy", "qwerty")

	req2, err := http.NewRequest(http.MethodGet, address+"/api/folders", nil)
	require.NoError(t, err, "cant create req")
	req2.Header.Add("Authorization", "Token "+tokenArseniy)
	resp2, err := client.Do(req2)
	require.NoError(t, err, "cant do request")
	defer resp2.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var folders2 ListFolders

	err = json.NewDecoder(resp2.Body).Decode(&folders2)
	require.NoError(t, err)

	require.Equal(t, 3, len(folders2.Folders))

	arr2 := []string{"anime", "magbitva", "genshin"}
	for i := range 3 {
		require.Contains(t, arr2, folders2.Folders[i].Name)
	}

}

func TestDeleteFolder(t *testing.T) {
	tokenDima := loginNoAdmin(t, "dima", "marvel")

	testscase := []struct {
		name      string
		folder_id string
		wantCode  int
	}{
		{
			name:      "OK_1",
			folder_id: "1",
			wantCode:  http.StatusOK,
		},
		{
			name:      "NoPermission",
			folder_id: "2",
			wantCode:  http.StatusForbidden,
		},
		{
			name:      "NotExist",
			folder_id: "1",
			wantCode:  http.StatusNotFound,
		},
	}

	for _, tc := range testscase {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodDelete, address+"/api/folders/"+tc.folder_id, nil)
			require.NoError(t, err, "cant create req")
			req.Header.Add("Authorization", "Token "+tokenDima)
			resp, err := client.Do(req)
			require.NoError(t, err, "cant do request")
			defer resp.Body.Close()

			require.Equal(t, tc.wantCode, resp.StatusCode)
		})
	}

}

func TestAddComics(t *testing.T) {
	token := login(t)
	update(token)
	time.Sleep(1 * time.Second)
	tokenDima := loginNoAdmin(t, "dima", "marvel")
	tokenArseniy := loginNoAdmin(t, "arseniy", "qwerty")
	testcase := []struct {
		name      string
		folder_id string
		comic_id  string
		wantCode  int
		dima      bool
		arseniy   bool
	}{
		{
			name:      "OK_1",
			folder_id: "3",
			comic_id:  "10",
			wantCode:  http.StatusOK,
			dima:      true,
		},
		{
			name:      "OK_2",
			folder_id: "3",
			comic_id:  "11",
			wantCode:  http.StatusOK,
			dima:      true,
		},
		{
			name:      "OK_3",
			folder_id: "4",
			comic_id:  "15",
			wantCode:  http.StatusOK,
			dima:      true,
		},
		{
			name:      "OK_4",
			folder_id: "4",
			comic_id:  "10",
			wantCode:  http.StatusOK,
			dima:      true,
		},
		{
			name:      "OK_5",
			folder_id: "4",
			comic_id:  "11",
			wantCode:  http.StatusOK,
			dima:      true,
		},
		{
			name:      "OK_6",
			folder_id: "5",
			comic_id:  "10",
			wantCode:  http.StatusOK,
			arseniy:   true,
		},
		{
			name:      "OK_7",
			folder_id: "6",
			comic_id:  "10",
			wantCode:  http.StatusOK,
			arseniy:   true,
		},
		{
			name:      "Conflict_1",
			folder_id: "4",
			comic_id:  "11",
			wantCode:  http.StatusConflict,
			dima:      true,
		},
		{
			name:      "Conflict_2",
			folder_id: "4",
			comic_id:  "10",
			wantCode:  http.StatusConflict,
			dima:      true,
		},
		{
			name:      "Conflict_3",
			folder_id: "3",
			comic_id:  "10",
			wantCode:  http.StatusConflict,
			dima:      true,
		},
		{
			name:      "Conflict_4",
			folder_id: "3",
			comic_id:  "11",
			wantCode:  http.StatusConflict,
			dima:      true,
		},
		{
			name:      "Conflict_5",
			folder_id: "5",
			comic_id:  "10",
			wantCode:  http.StatusConflict,
			arseniy:   true,
		},
		{
			name:      "NotExist",
			folder_id: "3",
			comic_id:  "10000",
			wantCode:  http.StatusNotFound,
			dima:      true,
		},
		{
			name:      "NoPermission_1",
			folder_id: "5",
			comic_id:  "10",
			wantCode:  http.StatusForbidden,
			dima:      true,
		},
		{
			name:      "NoPermission_2",
			folder_id: "3",
			comic_id:  "10",
			wantCode:  http.StatusForbidden,
			arseniy:   true,
		},
		{
			name:      "NoPermission_3",
			folder_id: "4",
			comic_id:  "10",
			wantCode:  http.StatusForbidden,
			arseniy:   true,
		},
		{
			name:      "NoPermission_4",
			folder_id: "6",
			comic_id:  "10",
			wantCode:  http.StatusForbidden,
			dima:      true,
		},
	}
	for _, tc := range testcase {
		t.Run(tc.name, func(t *testing.T) {
			if tc.dima {
				req1, err := http.NewRequest(http.MethodPost, address+"/api/folders/"+tc.folder_id+"/"+tc.comic_id, nil)
				require.NoError(t, err, "cant create req")
				req1.Header.Add("Authorization", "Token "+tokenDima)
				resp, err := client.Do(req1)
				require.NoError(t, err, "cant do request")
				defer resp.Body.Close()
				require.Equal(t, tc.wantCode, resp.StatusCode)
			}

			if tc.arseniy {
				req2, err := http.NewRequest(http.MethodPost, address+"/api/folders/"+tc.folder_id+"/"+tc.comic_id, nil)
				require.NoError(t, err, "cant create req")
				req2.Header.Add("Authorization", "Token "+tokenArseniy)
				resp, err := client.Do(req2)
				require.NoError(t, err, "cant do request")
				defer resp.Body.Close()
				require.Equal(t, tc.wantCode, resp.StatusCode)
			}
		})
	}
}

func TestAddComicWithoutAuth(t *testing.T) {
	req, err := http.NewRequest(http.MethodPost, address+"/api/folders/2/2", nil)
	require.NoError(t, err, "cant make request")

	resp, err := client.Do(req)
	require.NoError(t, err, "could not send login command")
	defer resp.Body.Close()

	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestListComicsWithoutAuth(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, address+"/api/folders/2/comics", nil)
	require.NoError(t, err, "cant make request")

	resp, err := client.Do(req)
	require.NoError(t, err, "could not send login command")
	defer resp.Body.Close()

	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

type Comic struct {
	ID  int    `json:"id"`
	URL string `json:"url"`
}

type ListComics struct {
	Comics []Comic `json:"comics"`
	Total  int     `json:"total"`
}

func TestListComics(t *testing.T) {
	tokenDima := loginNoAdmin(t, "dima", "marvel")
	tokenArseniy := loginNoAdmin(t, "arseniy", "qwerty")
	testcase := []struct {
		name      string
		folder_id string
		wantCode  int
		token     string
		data      []int
		Error     bool
	}{
		{
			name:      "Folder_Dima_3",
			folder_id: "3",
			wantCode:  http.StatusOK,
			token:     tokenDima,
			data:      []int{10, 11},
		},
		{
			name:      "Folder_Dima_4",
			folder_id: "4",
			wantCode:  http.StatusOK,
			token:     tokenDima,
			data:      []int{10, 11, 15},
		},
		{
			name:      "Folder_Arseniy_5",
			folder_id: "5",
			wantCode:  http.StatusOK,
			token:     tokenArseniy,
			data:      []int{10},
		},
		{
			name:      "Folder_Arseniy_6",
			folder_id: "6",
			wantCode:  http.StatusOK,
			token:     tokenArseniy,
			data:      []int{10},
		},
		{
			name:      "NoPermission_1",
			folder_id: "3",
			wantCode:  http.StatusForbidden,
			token:     tokenArseniy,
			Error:     true,
		},
		{
			name:      "NoPermission_2",
			folder_id: "4",
			wantCode:  http.StatusForbidden,
			token:     tokenArseniy,
			Error:     true,
		},
		{
			name:      "NoPermission_3",
			folder_id: "5",
			wantCode:  http.StatusForbidden,
			token:     tokenDima,
			Error:     true,
		},
		{
			name:      "NoPermission_4",
			folder_id: "6",
			wantCode:  http.StatusForbidden,
			token:     tokenDima,
			Error:     true,
		},
		{
			name:      "NotExist",
			folder_id: "100",
			wantCode:  http.StatusNotFound,
			token:     tokenDima,
			Error:     true,
		},
	}
	for _, tc := range testcase {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, address+"/api/folders/"+tc.folder_id+"/comics", nil)
			require.NoError(t, err, "cant create req")
			req.Header.Add("Authorization", "Token "+tc.token)
			resp, err := client.Do(req)
			require.NoError(t, err, "cant do req")
			defer resp.Body.Close()
			require.Equal(t, tc.wantCode, resp.StatusCode)

			if !tc.Error {
				var list ListComics
				err = json.NewDecoder(resp.Body).Decode(&list)
				require.NoError(t, err, "cant decode")
				for _, v := range list.Comics {
					require.Contains(t, tc.data, v.ID)
				}
			}
		})
	}

}

func TestDeleteComics(t *testing.T) {
	tokenDima := loginNoAdmin(t, "dima", "marvel")

	testscase := []struct {
		name      string
		folder_id string
		comic_id  string
		wantCode  int
	}{
		{
			name:      "OK_1",
			folder_id: "3",
			comic_id:  "10",
			wantCode:  http.StatusOK,
		},
		{
			name:      "OK_2",
			folder_id: "4",
			comic_id:  "15",
			wantCode:  http.StatusOK,
		},
		{
			name:      "OK_3",
			folder_id: "4",
			comic_id:  "11",
			wantCode:  http.StatusOK,
		},
		{
			name:      "OK_4",
			folder_id: "4",
			comic_id:  "10",
			wantCode:  http.StatusOK,
		},
		{
			name:      "NoPermission",
			folder_id: "5",
			comic_id:  "10",
			wantCode:  http.StatusForbidden,
		},
		{
			name:      "NotExist_1",
			folder_id: "1",
			comic_id:  "10",
			wantCode:  http.StatusNotFound,
		},
		{
			name:      "NotExist_2",
			folder_id: "100",
			comic_id:  "10",
			wantCode:  http.StatusNotFound,
		},
		{
			name:      "NotExist_3",
			folder_id: "3",
			comic_id:  "111",
			wantCode:  http.StatusNotFound,
		},
		{
			name:      "NotExist_4",
			folder_id: "3",
			comic_id:  "10",
			wantCode:  http.StatusNotFound,
		},
	}

	for _, tc := range testscase {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodDelete, address+"/api/folders/"+tc.folder_id+"/"+tc.comic_id, nil)
			require.NoError(t, err, "cant create req")
			req.Header.Add("Authorization", "Token "+tokenDima)
			resp, err := client.Do(req)
			require.NoError(t, err, "cant do request")
			defer resp.Body.Close()

			require.Equal(t, tc.wantCode, resp.StatusCode)
		})
	}
}
