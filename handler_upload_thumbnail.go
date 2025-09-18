package main

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/database"
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
		respondWithError(w, http.StatusInternalServerError, "failed to parse multipart form", err)
		return
	}

	multipartFile, multipartHeader, err := r.FormFile("thumbnail")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "unable to parse from file", err)
		return
	}
	defer multipartFile.Close()

	fileContentType := multipartHeader.Header.Get("Content-Type")

	rawFile, err := io.ReadAll(multipartFile)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed to read file", err)
		return
	}

	metadata, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "no video with the given id was found", err)
		return
	}
	if metadata.UserID != userID {
		respondWithError(w, http.StatusForbidden, "you have no access to that video", err)
		return
	}

	newThumbnail := thumbnail{
		data:      rawFile,
		mediaType: fileContentType,
	}
	videoThumbnails[videoID] = newThumbnail

	thumbnailUrl := "http://localhost:8091/api/thumbnails/" + videoIDString
	newVideo := database.Video{
		ID:           metadata.ID,
		CreatedAt:    metadata.CreatedAt,
		UpdatedAt:    time.Now(),
		ThumbnailURL: &thumbnailUrl,
		VideoURL:     nil,
		CreateVideoParams: database.CreateVideoParams{
			Title:       metadata.Title,
			Description: metadata.Description,
			UserID:      metadata.UserID,
		},
	}
	err = cfg.db.UpdateVideo(newVideo)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed to update metadata of the video", err)
		return
	}

	respondWithJSON(w, http.StatusOK, newVideo)
}
