package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/jung-kurt/gofpdf"
	"golang.org/x/net/html"
)

const LOGIN_URL = "https://lib.msk.misis.ru/elib/login.php"
const PAGE_URL = "https://lib.msk.misis.ru/elib/libs/view.php?id=%d&page=%d&type=large/fast"
const VIEW_URL = "https://lib.msk.misis.ru/elib/view.php?id=%d"

const file = "config.json"

type ElibClient struct {
	client *http.Client
}

func NewElibClient() *ElibClient {
	jar, err := cookiejar.New(nil)
	if err != nil {
		fmt.Println(err)
	}
	client := &http.Client{
		Jar: jar,
	}
	return &ElibClient{client: client}
}

func (ec *ElibClient) LoadCredentials() error {
	file := "config.json"
	_, err := os.Stat(file)
	if os.IsNotExist(err) {
		return ErrCredentialsConfgNotExist
	}

	f, err := os.Open(file)
	if err != nil {
		return ErrCannotLoadCredentials
	}

	defer f.Close()

	var data Credentials

	err = json.NewDecoder(f).Decode(&data)
	if err != nil {
		fmt.Println(fmt.Errorf("decode credentials: %w", err))
	}

	studentId := data.StudentId
	name := data.Name

	err = ec.Login(studentId, name)
	if err != nil {
		return ErrAuthFailed
	}

	return nil
}

func (ec *ElibClient) SaveCredentials(studentId int, name string) {
	_, err := os.Stat(file)
	if os.IsNotExist(err) {
		_, err := os.Create(file)
		if err != nil {
			fmt.Println(err)
		}
	}

	f, err := os.OpenFile(file, os.O_RDWR, 0644)
	if err != nil {
		fmt.Println(err)
	}

	defer f.Close()
	err = json.NewEncoder(f).Encode(Credentials{StudentId: studentId, Name: name})
	if err != nil {
		fmt.Println(err)
	}

}

func (ec *ElibClient) Login(studentId int, name string) error {
	url := "https://lib.msk.misis.ru/elib/login.php"
	method := "POST"

	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	_ = writer.WriteField("login", fmt.Sprintf("%d", studentId))
	_ = writer.WriteField("password", name)
	_ = writer.WriteField("language", "ru")
	err := writer.Close()
	if err != nil {
		fmt.Println(err)
		return err
	}

	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	res, err := ec.client.Do(req)
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println("logged in with status", res.StatusCode)
	defer res.Body.Close()

	ec.client.Jar.SetCookies(res.Request.URL, res.Cookies())

	return nil
}

func (ec *ElibClient) getPageCnt(r io.Reader) (int, error) {
	// read page count from option tag for selectiong avaliable pages
	doc, err := html.Parse(r)
	if err != nil {
		return 0, err
	}
	var lastCnt int
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "option" {
			lastCnt, err = strconv.Atoi(n.Attr[0].Val)
			if err != nil {
				return
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)
	return lastCnt, nil
}

func (ec *ElibClient) GetPageCnt(bookId int) (int, error) {
	// get page count for book from view page
	res, err := ec.client.Get(fmt.Sprintf(VIEW_URL, bookId))
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()
	cnt, err := ec.getPageCnt(res.Body)
	if err != nil {
		return 0, err
	}
	fmt.Println("page count:", cnt)
	return cnt, nil
}

func (ec *ElibClient) downloadPage(bookId int, page int, pageCh chan DocumentPage, wg *sync.WaitGroup) error {
	// download page from elib system and save it to file

	defer wg.Done()
	res, err := ec.client.Get(fmt.Sprintf(PAGE_URL, bookId, page))
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer res.Body.Close()

	pageData := DocumentPage{bookId: bookId, number: page, image: bytes.Buffer{}}
	_, err = pageData.image.ReadFrom(res.Body)
	if err != nil {
		fmt.Println(err)
		return err
	}
	pageCh <- pageData
	return nil
}

func (ec *ElibClient) DownloadDocument(bookId int) error {
	startTime := time.Now()

	pageCnt, err := ec.GetPageCnt(bookId)
	var wg sync.WaitGroup
	if err != nil {
		panic(err)
	}
	pageCh := make(chan DocumentPage, pageCnt)
	fmt.Println("downloading", pageCnt, "pages")
	for cur := 0; cur < pageCnt; cur++ {
		wg.Add(1)
		go ec.downloadPage(bookId, cur, pageCh, &wg)
	}

	wg.Wait()
	fmt.Println("all pages downloaded")
	pages := make([]DocumentPage, pageCnt)

	for i := 0; i < pageCnt; i++ {
		page := <-pageCh
		pages[page.number] = page
	}

	close(pageCh)

	pdf := gofpdf.New("P", "mm", "A4", "")
	for _, page := range pages {
		// Save the image to a temporary file
		tmpFile, err := os.CreateTemp("", "page")
		if err != nil {
			fmt.Println(err)
			return err
		}
		_, err = tmpFile.Write(page.image.Bytes())
		if err != nil {
			fmt.Println(err)
			return err
		}

		// Add a new page to the PDF and add the image to the page
		pdf.AddPage()
		pdf.ImageOptions(tmpFile.Name(), 15, 15, 180, 0, false, gofpdf.ImageOptions{ImageType: "JPG", ReadDpi: true}, 0, "")

		// Close and remove the temporary file
		tmpFile.Close()
		os.Remove(tmpFile.Name())
	}
	err = pdf.OutputFileAndClose(fmt.Sprintf("book_%d.pdf", bookId))
	if err != nil {
		panic(err)
	}
	fmt.Println("downloaded in", time.Since(startTime))
	return nil
}

func (ec *ElibClient) BrowseFolder(folderId, start, limit, retry int) {
	fmt.Println("Browsing folder:", folderId, start, limit)

	var jsonData ReadInFolderResponse
	if _, err := os.Stat(fmt.Sprintf("./folders/folder_%d.json", folderId)); err == nil {

		fmt.Printf("folder %d already exists reading from file", folderId)
		file, err := os.Open(fmt.Sprintf("./folders/folder_%d.json", folderId))
		if err != nil {
			fmt.Println(err)
			return
		}
		defer file.Close()
		err = json.NewDecoder(file).Decode(&jsonData)
		if err != nil {
			fmt.Println(err)
			return
		}

	} else {
		url := "https://lib.msk.misis.ru/elib/libs/api.php"
		method := "POST"

		payload := &bytes.Buffer{}
		writer := multipart.NewWriter(payload)
		_ = writer.WriteField("action", "ReadInFolder")
		_ = writer.WriteField("folder_id", fmt.Sprintf("%d", folderId))
		_ = writer.WriteField("start", fmt.Sprintf("%d", start))
		_ = writer.WriteField("limit", fmt.Sprintf("%d", limit))
		_ = writer.WriteField("sort_id", "0")
		_ = writer.WriteField("orderby", "asc")
		err := writer.Close()
		if err != nil {
			fmt.Println(err)
			return
		}
		req, err := http.NewRequest(method, url, payload)
		if err != nil {
			fmt.Println(err)
			return
		}
		req.Header.Set("Content-Type", writer.FormDataContentType())

		res, err := ec.client.Do(req)
		if err != nil {
			fmt.Println(err)
			return
		}

		defer res.Body.Close()
		err = json.NewDecoder(res.Body).Decode(&jsonData)

		if err != nil {

			if retry > 10 {
				fmt.Println("exceded maximum retries")
				return
			}

			if res.StatusCode == 504 {
				// sleep for 0.5 seconds and retry
				fmt.Println("retrying", folderId, retry)
				time.Sleep(500 * time.Millisecond)
				go ec.BrowseFolder(folderId, 0, 100, retry+1)
			} else if res.StatusCode == 503 {
				fmt.Println("503, waiting for 2 seconds")
				time.Sleep(2000 * time.Millisecond)

				go ec.BrowseFolder(folderId, 0, 100, retry+1)
			} else {
				fmt.Println("jsonData err", res.StatusCode, res)
				return
			}

			return
		}

		file := fmt.Sprintf("./folders/folder_%d.json", folderId)
		f, err := os.Create(file)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer f.Close()
		err = json.NewEncoder(f).Encode(jsonData)
		if err != nil {
			fmt.Println(err)
			return
		}

	}

	if len(jsonData.Data.Folders) > 0 {
		for _, folder := range jsonData.Data.Folders {
			id, err := strconv.Atoi(folder.ID)
			if err != nil {
				fmt.Println(err)
				continue
			}
			time.Sleep(200 * time.Millisecond)
			go ec.BrowseFolder(id, 0, 100, 0)
		}
	}

}

func (ec *ElibClient) ParseLib() {

	documentIdCh := make(chan DocuemntResponse)

	rootFolderId := 1
	go ec.BrowseFolder(rootFolderId, 0, 100, 0)
	for document := range documentIdCh {

		fmt.Println(document)

		file := fmt.Sprintf("document_%s.json", document.ID)
		f, err := os.Create(file)
		if err != nil {
			fmt.Println(err)
			continue
		}
		err = json.NewEncoder(f).Encode(document)
		if err != nil {
			fmt.Println(err)
			continue
		}
		f.Close()

	}

}

func (ec *ElibClient) Close() {
	ec.client.CloseIdleConnections()
}
