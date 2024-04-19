package reports

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	firebase "firebase.google.com/go"
	"google.golang.org/api/option"
)

func UploadToFirebaseStorage(pdfFilePath string) (string, error) {
	ctx := context.Background()
	opt := option.WithCredentialsFile("firebase.json")
	config := &firebase.Config{
		StorageBucket: os.Getenv("FIREBASE_STORAGE"),
		ProjectID:     "stndup-df313",
	}
	app, err := firebase.NewApp(ctx, config, opt)
	if err != nil {
		return "", fmt.Errorf("error initializing firebase app: %v", err)
	}

	client, err := app.Storage(ctx)
	if err != nil {
		return "", fmt.Errorf("error initializing firebase storage client: %v", err)
	}

	file, err := os.Open(pdfFilePath)
	if err != nil {
		return "", fmt.Errorf("error opening PDF file: %v", err)
	}
	defer file.Close()

	bucket, err := client.DefaultBucket()
	if err != nil {
		return "", fmt.Errorf("error getting default storage bucket: %v", err)
	}

	obj := bucket.Object(filepath.Base(pdfFilePath))
	wc := obj.NewWriter(ctx)
	if _, err := io.Copy(wc, file); err != nil {
		return "", fmt.Errorf("error uploading PDF file to storage: %v", err)
	}
	if err := wc.Close(); err != nil {
		return "", fmt.Errorf("error closing writer: %v", err)
	}

	attrs, err := obj.Attrs(ctx)
	if err != nil {
		return "", fmt.Errorf("error getting object attributes: %v", err)
	}

	downloadURL := attrs.MediaLink
	return downloadURL, nil
}
