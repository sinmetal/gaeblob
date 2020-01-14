package backend

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"cloud.google.com/go/storage"
	"github.com/google/uuid"
	"golang.org/x/oauth2/google"
)

const bucket = "sinmetal-lab-blob"

type SignedURL struct {
	Key            string `json:"key"`
	Bucket         string `json:"bucket"`
	ContentType    string `json:"contentType"`
	GoogleAccessID string `json:"googleAccessId"`
	ACL            string `json:"acl"`
	Policy         string `json:"policy"`
	Signature      string `json:"signature"`
}

func UploadURLHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ct := r.FormValue("contentType")
	if ct == "" {
		http.Error(w, "contentType must be set", http.StatusBadRequest)
		return
	}

	creds, err := google.FindDefaultCredentials(ctx, storage.ScopeReadOnly)
	if err != nil {
		fmt.Printf("failed google.FindDefaultCredentials err=%+v\n", err)
		http.Error(w, "failed google.FindDefaultCredentials", http.StatusInternalServerError)
		return
	}
	conf, err := google.JWTConfigFromJSON(creds.JSON, storage.ScopeReadWrite)
	if err != nil {
		fmt.Printf("failed google.JWTConfigFromJSON err=%+v\n", err)
		http.Error(w, "failed google.JWTConfigFromJSON", http.StatusInternalServerError)
		return
	}

	object := uuid.New().String()
	u, err := storage.SignedURL(bucket, object, &storage.SignedURLOptions{
		GoogleAccessID: conf.Email,
		PrivateKey:     conf.PrivateKey,
		Method:         http.MethodPut,
		ContentType:    ct,
		Expires:        time.Now().Add(10 * time.Minute),
	})
	if err != nil {
		fmt.Printf("failed storage.SignedURL() bucket=%v,object=%v err=%+v\n", bucket, object, err)
		http.Error(w, "failed storage.SignedURL", http.StatusInternalServerError)
		return
	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte(fmt.Sprintf(`{"url":"%s"}`, u)))
	if err != nil {
		fmt.Printf("failed write to response err=%+v\n", err)
	}
}

func DownloadURLHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	object := r.FormValue("object")
	fmt.Printf("object is %v\n", object)

	creds, err := google.FindDefaultCredentials(ctx, storage.ScopeReadOnly)
	if err != nil {
		fmt.Printf("failed google.FindDefaultCredentials err=%+v\n", err)
		http.Error(w, "failed google.FindDefaultCredentials", http.StatusInternalServerError)
		return
	}
	conf, err := google.JWTConfigFromJSON(creds.JSON, storage.ScopeReadOnly)
	if err != nil {
		fmt.Printf("failed google.JWTConfigFromJSON err=%+v\n", err)
		http.Error(w, "failed google.JWTConfigFromJSON", http.StatusInternalServerError)
		return
	}

	u, err := storage.SignedURL(bucket, object, &storage.SignedURLOptions{
		GoogleAccessID: conf.Email,
		PrivateKey:     conf.PrivateKey,
		Method:         http.MethodGet,
		Expires:        time.Now().Add(10 * time.Minute),
	})
	if err != nil {
		fmt.Printf("failed storage.SignedURL() bucket=%v,object=%v err=%+v\n", bucket, object, err)
		http.Error(w, "failed storage.SignedURL", http.StatusInternalServerError)
		return
	}

	v := url.Values{}
	v.Set("response-content-disposition", fmt.Sprintf(`attachment; filename=%s`, "hogeFile"))

	u = fmt.Sprintf("%s&%s", u, v.Encode())

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte(fmt.Sprintf(`{"url":"%s"}`, u)))
	if err != nil {
		fmt.Printf("failed write to response err=%+v\n", err)
	}
}
