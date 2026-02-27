package google

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/masudur-rahman/expense-tracker-bot/infra/logr"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

const (
	dbFileName   = "expense-tracker.db"
	dbFolderName = ".expense-tracker"
)

func DatabasePath() string {
	if err := os.MkdirAll(dbFolderName, os.ModePerm); err != nil {
		log.Fatalln(err)
	}
	return filepath.Join(dbFolderName, dbFileName)
}

func getDriveService() (*drive.Service, error) {
	creds, err := google.FindDefaultCredentials(context.Background(), drive.DriveScope)
	if err != nil {
		return nil, err
	}

	return drive.NewService(context.Background(), option.WithCredentials(creds))
}

func readFileFromDrive(svc *drive.Service, upstreamFilePath string) ([]byte, error) {
	ff := strings.Split(upstreamFilePath, "/")

	folderID, err := getFolderID(svc, ff[0])
	if err != nil {
		return nil, err
	}

	fileQuery := fmt.Sprintf("name='%s' and '%s' in parents", ff[1], folderID)
	fileList, err := svc.Files.List().Q(fileQuery).Do()
	if err != nil {
		return nil, err
	}

	if len(fileList.Files) == 0 {
		return nil, fmt.Errorf("file '%s' not found in folder '%s'", ff[1], ff[0])
	}

	fileID := fileList.Files[0].Id

	resp, err := svc.Files.Get(fileID).Download()
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func uploadFileToDrive(svc *drive.Service, upstreamFilePath string, localFilePath string) error {
	ff := strings.Split(upstreamFilePath, "/")
	folderID, err := getFolderID(svc, ff[0])
	if err != nil {
		return err
	}

	fileMetadata := &drive.File{
		Name:     filepath.Base(ff[1]),
		Parents:  []string{folderID},
		MimeType: "application/octet-stream",
	}

	file, err := os.Open(localFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	gf, err := svc.Files.Create(fileMetadata).Media(file).Do()
	if err != nil {
		return err
	}

	//_, err = svc.Files.Update(existingFiles.Files[0].Id, fileMetadata).Media(file).Do()
	//return err

	return removePreviousFiles(svc, ff[1], gf.Id, folderID)
}

func removePreviousFiles(svc *drive.Service, fileName, newFileID string, folderID string) error {
	fileQuery := fmt.Sprintf("'%s' in parents and name='%s'", folderID, fileName)
	fileList, err := svc.Files.List().Q(fileQuery).Do()
	if err != nil {
		return err
	}
	for _, file := range fileList.Files {
		if file.Id != newFileID {
			if err = svc.Files.Delete(file.Id).Do(); err != nil {
				return err
			}
		}
	}
	return nil
}

func getFolderID(svc *drive.Service, folderName string) (string, error) {
	folderQuery := fmt.Sprintf("name='%s' and mimeType='application/vnd.google-apps.folder'", folderName)
	folderList, err := svc.Files.List().Q(folderQuery).Do()
	if err != nil {
		return "", err
	}

	if len(folderList.Files) == 0 {
		parentFolder := &drive.File{
			Name:     folderName,
			MimeType: "application/vnd.google-apps.folder",
		}
		parentFolder, err = svc.Files.Create(parentFolder).Do()
		if err != nil {
			return "", err
		}
		return parentFolder.Id, nil
	}

	return folderList.Files[0].Id, nil
}

func SyncDatabaseFromDrive() error {
	svc, err := getDriveService()
	if err != nil {
		return err
	}

	data, err := readFileFromDrive(svc, DatabasePath())
	if err != nil {
		return err
	}

	return os.WriteFile(DatabasePath(), data, 0666)
}

func SyncDatabaseToDrive() error {
	svc, err := getDriveService()
	if err != nil {
		return err
	}

	dbPath := DatabasePath()
	return uploadFileToDrive(svc, dbPath, dbPath)
}

func SyncDatabaseToDrivePeriodically(interval time.Duration) {
	if interval == time.Duration(0) {
		interval = time.Hour
	}

	if err := SyncDatabaseToDrive(); err != nil {
		logr.DefaultLogger.Errorw("Sync database to drive failed", "error", err.Error())
		return
	}
	logr.DefaultLogger.Infof("SQLite database synced to google drive")

	ticker := time.NewTicker(interval)
	for range ticker.C {
		if err := SyncDatabaseToDrive(); err != nil {
			logr.DefaultLogger.Errorw("Sync database to drive failed", "error", err.Error())
			return
		}
		logr.DefaultLogger.Infof("SQLite database synced to google drive")
	}
}
