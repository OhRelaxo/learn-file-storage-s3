package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadThumbnail(w http.ResponseWriter, r *http.Request) {
	videoIDString := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}

	fmt.Println("uploading thumbnail for video", videoID, "by user", userID)

	const maxMemory = 10 << 20

	err = r.ParseMultipartForm(maxMemory)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to parse multipart form", err)
		return
	}

	multipartFile, multipartHeader, err := r.FormFile("thumbnail")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to parse from file", err)
		return
	}
	defer multipartFile.Close()

	fileContentType := multipartHeader.Header.Get("Content-Type")
	if fileContentType == "" {
		respondWithError(w, http.StatusBadRequest, "Missing Content-Type for thumbnail", nil)
		return
	}

	video, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't find video", err)
		return
	}
	if video.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "Not authorized to update this video", nil)
		return
	}

	fileExtension := getFileExtension(fileContentType)
	assetsPath := getAssetsPath(videoID, fileExtension, cfg.assetsRoot)

	newFile, err := os.Create(assetsPath)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create file on server", err)
		return
	}

	_, err = io.Copy(newFile, multipartFile)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error while coping data to file", err)
		return
	}

	thumbnailUrl := fmt.Sprintf("http://localhost:%v/assets/%v.%v", cfg.port, videoID, fileExtension)
	video.ThumbnailURL = &thumbnailUrl
	video.UpdatedAt = time.Now()
	err = cfg.db.UpdateVideo(video)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to update video", err)
		return
	}

	respondWithJSON(w, http.StatusOK, video)
}
