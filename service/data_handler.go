package service

import (
	"ListCharacterGI/cfg"
	"ListCharacterGI/model/entity"
	"ListCharacterGI/model/request"
	"ListCharacterGI/model/respErr"
	"ListCharacterGI/repository"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"net/http"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type Handler struct {
	GenshinRepository repository.GenshinRepository
}

func NewGenshinService(genshinRepository repository.GenshinRepository) *Handler {
	return &Handler{
		GenshinRepository: genshinRepository,
	}
}

// Fungsi Register
func (h *Handler) Register(ctx *gin.Context) {
	var user entity.User

	// binding request body ke struct user
	if err := ctx.ShouldBindJSON(&user); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, respErr.ErrorResponse{
			Message: err.Error(),
			Status:  http.StatusBadRequest,
		})
		return
	}
	// Validasi alamat email
	if !strings.HasSuffix(user.Email, "@gmail.com") {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, respErr.Error{
			Error: "Invalid email format. Email must be a @gmail.com address.",
		})
		return
	}

	// cek apakah username sudah ada di database
	existingUser, err := h.GenshinRepository.GetUserByUsernameOrEmail(user.Username, user.Email)
	if existingUser != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, respErr.ErrorResponse{
			Message: "Username, or email already exist",
			Status:  http.StatusBadRequest,
		})
		return
	}

	// hash password pengguna sebelum disimpan ke database
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, respErr.ErrorResponse{
			Message: "Failed to hash Password",
			Status:  http.StatusInternalServerError,
		})
		return
	}

	// simpan pengguna ke database
	newUser := &entity.User{
		Username: user.Username,
		Password: string(hashedPassword),
		Email:    user.Email,
		Role:     user.Role,
	}
	err = h.GenshinRepository.CreateUser(newUser)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, respErr.ErrorResponse{
			Message: err.Error(),
			Status:  http.StatusInternalServerError,
		})
		return
	}

	// mengembalikan pesan berhasil sebagai response
	ctx.JSON(http.StatusOK, gin.H{"message": "User created successfully"})
}

func (h *Handler) Login(ctx *gin.Context) {
	var userLogin entity.UserLogin

	// binding request body ke struct UserLogin
	if err := ctx.ShouldBindJSON(&userLogin); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, respErr.ErrorResponse{
			Message: "Invalid request Body",
			Status:  http.StatusBadRequest,
		})
		return
	}

	// Cek apakah pengguna ada di database berdasarkan username atau email
	storedUser, err := h.GenshinRepository.GetUserByUsernameOrEmail(userLogin.Username, userLogin.Email)
	if err != nil || storedUser == nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, respErr.ErrorResponse{
			Message: "Invalid Username or Password",
			Status:  http.StatusUnauthorized,
		})
		return
	}

	// Membandingkan password yang dimasukkan dengan hash password di database
	err = bcrypt.CompareHashAndPassword([]byte(storedUser.Password), []byte(userLogin.Password))
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, respErr.ErrorResponse{
			Message: "Invalid Username or Password",
			Status:  http.StatusUnauthorized,
		})
		return
	}

	// Membuat token
	token, err := cfg.CreateToken(storedUser.Username, storedUser.ID, storedUser.Role)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, respErr.ErrorResponse{
			Message: "Failed to generate Token",
			Status:  http.StatusInternalServerError,
		})
		return
	}

	response := request.LoginResponse{
		Message: fmt.Sprintf("Hello %s! You are not logged in.", userLogin.Username),
		Token:   token,
		UserID:  int(storedUser.ID),
	}

	ctx.JSON(http.StatusOK, response)
}

func (h *Handler) Access(ctx *gin.Context) {
	// ambil username dari konteks
	username, ok := ctx.Get("username")
	userID, _ := ctx.Get("user_id")
	if !ok {
		// jika tidak ada username di dalam konteks, berarti pengguna belum terautentikasi
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, respErr.Error{
			Error: "Unauthorized",
		})
		return
	}

	// kirim pesan hello ke pengguna
	ctx.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Hello %s!", username),
		"user_id": userID,
	})
}

// Handler untuk menampilkan semua data pengguna
func (h *Handler) ViewAllUsers(ctx *gin.Context) {
	// Dapatkan data pengguna dari basis data
	users, err := h.GenshinRepository.GetAllUsers()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, respErr.ErrorResponse{
			Message: "Internal Server Error",
			Status:  http.StatusInternalServerError,
		})
		return
	}

	// Tampilkan data pengguna dalam format respons JSON
	ctx.JSON(http.StatusOK, users)
}

// Handler untuk menghapus pengguna dengan peran "user"
func (h *Handler) DeleteUser(ctx *gin.Context) {
	// Ambil user_id dari parameter URL
	userIDStr := ctx.Param("user_id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, respErr.ErrorResponse{
			Message: "Invalid user_id",
			Status:  http.StatusBadRequest,
		})
		return
	}

	// Hapus pengguna dengan peran "user" dari basis data
	err = h.GenshinRepository.DeleteUserByIDAndRole(userID, "user")
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, respErr.ErrorResponse{
			Message: "Internal Server Error",
			Status:  http.StatusInternalServerError,
		})
		return
	}

	// Tampilkan respons sukses
	ctx.JSON(http.StatusOK, request.SuccessMessage{
		Status:  http.StatusOK,
		Message: "User deleted successfully",
	})
}

func (h *Handler) HandlerGetAll(ctx *gin.Context) {
	lod, err := h.GenshinRepository.GetAllCharacters()
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, &respErr.ErrorResponse{
			Message: err.Error(),
			Status:  http.StatusInternalServerError,
		})
		return
	}

	logrus.Info(http.StatusOK, " Success Get All Data")
	ctx.AbortWithStatusJSON(http.StatusOK, request.ResponseToGetAll{
		Message: "Success Get All",
		Data:    len(lod),
		MHS:     lod,
	})

}
func (h *Handler) HandlerCreate(ctx *gin.Context) {
	data := new(request.ListCreateRequest)
	if err := ctx.ShouldBindJSON(data); err != nil {
		logrus.Error(err)
		ctx.JSON(http.StatusBadRequest, respErr.ErrorResponse{
			Message: err.Error(),
			Status:  http.StatusBadRequest,
		})
		return
	}

	// Get the user ID from the token
	userID, _ := ctx.Get("user_id")
	if userID == nil {
		ctx.JSON(http.StatusUnauthorized, respErr.ErrorResponse{
			Message: "User not authenticated",
			Status:  http.StatusUnauthorized,
		})
		return
	}

	// Cast the userID to int64
	userIDInt64, ok := userID.(int64)
	if !ok {
		ctx.JSON(http.StatusBadRequest, respErr.ErrorResponse{
			Message: "Invalid user_id",
			Status:  http.StatusBadRequest,
		})
		return
	}

	// Set the user ID in the CreateRequest
	data.UserID = userIDInt64

	// Check if data with the same name
	existingData, err := h.GenshinRepository.GetCharacterByName(data.Name)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, respErr.ErrorResponse{
			Message: "Internal Server Error",
			Status:  http.StatusInternalServerError,
		})
		return
	}
	if existingData != nil {
		ctx.JSON(http.StatusBadRequest, respErr.ErrorResponse{
			Message: "Data with the same name exists",
			Status:  http.StatusBadRequest,
		})
		return
	}

	//
	newData := &entity.Characters{
		UserID:     data.UserID,
		Name:       data.Name,
		Age:        data.Age,
		Address:    data.Address,
		WeaponType: data.WeaponType,
		Element:    data.Element,
		StarRating: data.StarRating,
	}

	createdData, errCreate := h.GenshinRepository.Create(newData)
	if errCreate != nil {
		ctx.JSON(http.StatusInternalServerError, respErr.ErrorResponse{
			Message: "Internal Server Error",
			Status:  http.StatusInternalServerError,
		})
		return
	}

	ctx.JSON(http.StatusOK, request.ListResponse{
		Status:  http.StatusOK,
		Message: "New Character Created",
		Data:    *createdData,
	})
}

func (h *Handler) HandlerGetByID(ctx *gin.Context) {
	userId := ctx.Param("id")
	Id, err := strconv.ParseInt(userId, 10, 64)
	if err != nil {
		logrus.Error(err)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, respErr.ErrorResponse{
			Message: "Bad request",
			Status:  http.StatusBadRequest,
		})
		return
	}
	list, err := h.GenshinRepository.GetByID(Id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logrus.Errorf("Character not found: %v", err)
			ctx.AbortWithStatusJSON(http.StatusNotFound, respErr.ErrorResponse{
				Message: "Character Not Found",
				Status:  http.StatusNotFound,
			})
			return
		}
		logrus.Errorf("Internal server error: %v", err)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, respErr.ErrorResponse{
			Message: err.Error(),
			Status:  http.StatusInternalServerError,
		})
		return
	}

	logrus.Info(http.StatusOK, " Success Get By ID")
	ctx.JSON(http.StatusOK, request.ListResponse{
		Status:  http.StatusOK,
		Message: "Success Get Id",
		Data:    *list,
	})
}

func (h *Handler) HandlerUpdate(ctx *gin.Context) {
	var requestData request.ListUpdateRequest
	if err := ctx.ShouldBindJSON(&requestData); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, respErr.ErrorResponse{
			Message: err.Error(),
			Status:  http.StatusBadRequest,
		})
		return
	}
	//characterId := requestData.ID
	characterId := ctx.Param("id")
	Id, err := strconv.ParseInt(characterId, 10, 64)
	if err != nil {
		logrus.Error(err)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, respErr.ErrorResponse{
			Message: "parse ID error",
			Status:  http.StatusBadRequest,
		})
		return
	}

	existingCharacter, err := h.GenshinRepository.GetByID(Id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logrus.Errorf("Character not found: %v", err)
			ctx.AbortWithStatusJSON(http.StatusNotFound, respErr.ErrorResponse{
				Message: "Character Not Found",
				Status:  http.StatusNotFound,
			})
			return
		}
		logrus.Errorf("Internal server error: %v", err)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, respErr.ErrorResponse{
			Message: err.Error(),
			Status:  http.StatusInternalServerError,
		})
		return
	}

	// Update the fields if provided in the request
	if requestData.Name != "" {
		existingCharacter.Name = requestData.Name
	}
	if requestData.Element != "" {
		existingCharacter.Element = requestData.Element
	}
	if requestData.Age != 0 {
		existingCharacter.Age = requestData.Age
	}
	if requestData.StarRating != "" {
		existingCharacter.StarRating = requestData.StarRating
	}
	if requestData.Address != "" {
		existingCharacter.Address = requestData.Address
	}

	updatedCharacter, err := h.GenshinRepository.Update(existingCharacter)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, respErr.ErrorResponse{
			Message: "err.Error()",
			Status:  http.StatusInternalServerError,
		})
		return
	}

	logrus.Info(http.StatusOK, " Success Update data")
	ctx.JSON(http.StatusOK, request.UpdateResponse{
		Status:  http.StatusOK,
		Message: "Success Update data",
		MHS:     updatedCharacter,
	})

}
func (h *Handler) HandlerDelete(ctx *gin.Context) {

	ListId := ctx.Param("id")
	Id, err := strconv.ParseInt(ListId, 10, 64)
	if err != nil {
		logrus.Error(err)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, respErr.ErrorResponse{
			Message: "Parse ID Error",
			Status:  http.StatusBadRequest,
		})
		return
	}

	// Delete the Data with the specified mhsID and userID
	_, err = h.GenshinRepository.Delete(Id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logrus.Errorf("Character not found: %v", err)
			ctx.AbortWithStatusJSON(http.StatusNotFound, respErr.ErrorResponse{
				Message: "Character Not Found",
				Status:  http.StatusNotFound,
			})
			return
		}
		logrus.Errorf("Internal server error: %v", err)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, respErr.ErrorResponse{
			Message: "Internal Server Error",
			Status:  http.StatusInternalServerError,
		})
		return
	}

	logrus.Info(http.StatusOK, " Success DELETE")
	ctx.JSON(http.StatusOK, request.DeleteResponse{
		Status:  http.StatusOK,
		Message: "Success Delete",
	})
}

func (h *Handler) UploadFileS3AtchHandler(ctx *gin.Context) {
	IDStr := ctx.Param("id")
	Id, err := strconv.ParseInt(IDStr, 10, 64)
	if err != nil {
		logrus.Error(err)
		ctx.JSON(http.StatusBadRequest, respErr.ErrorResponse{
			Message: err.Error(),
			Status:  http.StatusBadRequest,
		})
		return
	}

	// Check if the Data with the given ID exists
	mhs, err := h.GenshinRepository.GetByID(Id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, respErr.ErrorResponse{
			Message: "Todo not found",
			Status:  http.StatusNotFound,
		})
		return
	}

	file, err := ctx.FormFile("file")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, respErr.ErrorResponse{
			Message: "No File Upload",
			Status:  http.StatusBadRequest,
		})
		return
	}

	// Check file type yang boleh cuman jpg jpeg png webp
	allowedExtensions := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".webp": true,
	}
	ext := filepath.Ext(file.Filename)
	if !allowedExtensions[ext] {
		ctx.JSON(http.StatusBadRequest, respErr.ErrorResponse{
			Message: "error File not allowed type",
			Status:  http.StatusBadRequest,
		})
		return
	}

	// Use the GenshinRepository to upload the file to S3
	attachment, err := h.GenshinRepository.UploadFileS3Atch(file, Id)
	if err != nil {
		// Periksa apakah error merupakan "Data not found" atau bukan
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Jika error disebabkan oleh record not found, kirim respons 404
			ctx.JSON(http.StatusNotFound, respErr.ErrorResponse{
				Message: "data Character not found",
				Status:  http.StatusNotFound,
			})
		} else {
			// Jika error bukan karena record not found, kirim respons 500
			ctx.JSON(http.StatusInternalServerError, respErr.ErrorResponse{
				Message: err.Error(),
				Status:  http.StatusInternalServerError,
			})
			logrus.Error(err)
		}
		return
	}

	// Update the list_genshin Attachments field with the new attachment
	mhs.Attachments = append(mhs.Attachments, *attachment)

	// Create an attachment record in the database
	err = h.GenshinRepository.UpdateWithAttachments(mhs)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, respErr.ErrorResponse{
			Message: "Failed to update Todo with attachments",
			Status:  http.StatusInternalServerError,
		})
		return
	}

	ctx.JSON(http.StatusOK, request.SuccessMessage{
		Message: "File uploaded and attachment created successfully",
		Data:    attachment,
		Status:  http.StatusOK,
	})
}

func (h *Handler) UploadFileS3BucketsHandler(ctx *gin.Context) {
	file, err := ctx.FormFile("file")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "No File Upload",
		})
		return
	}

	src, err := file.Open()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to open file",
		})
		return
	}
	defer src.Close()

	// Use the mhsRepository to upload the file to S3
	publicURL, err := h.GenshinRepository.UploadFileS3Buckets(src, file.Filename)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to upload file to S3",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "File uploaded to S3 successfully",
		"url":     *publicURL,
	})

}

func (h *Handler) UploadLocalAtchHandler(ctx *gin.Context) {
	IDStr := ctx.Param("id")
	Id, err := strconv.ParseInt(IDStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, respErr.ErrorResponse{
			Message: "Invalid Data ID",
			Status:  http.StatusBadRequest,
		})
		return
	}

	// Check if the Data with given ID exists
	gi, err := h.GenshinRepository.GetByID(Id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, respErr.ErrorResponse{
			Message: "data Character not found",
			Status:  http.StatusBadRequest,
		})
		return
	}

	file, err := ctx.FormFile("file")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, respErr.ErrorResponse{
			Message: "No FIle Upload",
			Status:  http.StatusBadRequest,
		})
		return
	}

	// Check file
	allowedExtensions := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".webp": true,
	}
	ext := filepath.Ext(file.Filename)
	if !allowedExtensions[ext] {
		ctx.JSON(http.StatusBadRequest, respErr.ErrorResponse{
			Message: "error not allowed type",
			Status:  http.StatusBadRequest,
		})
		return
	}

	// use
	attachment, err := h.GenshinRepository.UploadFileLocalAtch(file, Id)
	if err != nil {
		// Periksa apakah error merupakan "data Character not found" atau bukan
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Jika error disebabkan oleh record not found, kirim respons 404
			ctx.JSON(http.StatusNotFound, respErr.ErrorResponse{
				Message: "data Character not found",
				Status:  http.StatusNotFound,
			})
		} else {
			// Jika error bukan karena record not found, kirim respons 500
			ctx.JSON(http.StatusInternalServerError, respErr.ErrorResponse{
				Message: err.Error(),
				Status:  http.StatusInternalServerError,
			})
			logrus.Error(err)
		}
		return
	}

	// Update the data Attachments field with the new attachment
	gi.Attachments = append(gi.Attachments, *attachment)

	// Save the updated data to the database
	err = h.GenshinRepository.UpdatetoAtch(gi)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, respErr.ErrorResponse{
			Message: err.Error(),
			Status:  http.StatusInternalServerError,
		})
		return
	}

	ctx.JSON(http.StatusOK, request.SuccessMessage{
		Status:  http.StatusOK,
		Message: "FIle Uploaded and attachment created successfully",
		Data:    attachment,
	})
}

func (h *Handler) DownloadAttachmentHandler(ctx *gin.Context) {
	attachmentIDStr := ctx.Param("id")
	attachmentID, err := strconv.ParseInt(attachmentIDStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, respErr.ErrorResponse{
			Message: "Invalid Attachment ID",
			Status:  http.StatusBadRequest,
		})
		return
	}

	// Get the attachment from the database using the attachmentID
	attachment, err := h.GenshinRepository.GetAttachmentByID(attachmentID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, respErr.ErrorResponse{
			Message: "Attachment not found",
			Status:  http.StatusNotFound,
		})
		return
	}

	relativePath := attachment.Path
	logrus.Info(relativePath)

	// Generate a pre-signed URL for downloading the attachment from S3
	presignedURL, err := h.GenshinRepository.GeneratePresignedURL("bucketwithrey", relativePath)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, respErr.ErrorResponse{
			Message: "Failed to generate pre-signed URL",
			Status:  http.StatusInternalServerError,
		})
		logrus.Info(presignedURL)
		return
	}

	// Return the pre-signed URL to the client
	ctx.JSON(http.StatusOK, gin.H{
		"presigned_url": presignedURL,
	})
}

func (h *Handler) SearchHandler(ctx *gin.Context) {

	searchQuery := ctx.Query("search")
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(ctx.DefaultQuery("per_page", "5"))

	dataGI, total, err := h.GenshinRepository.SearchCharacterByUser(searchQuery, page, perPage)
	if err != nil {
		logrus.Errorf("Error during character search: %v", err)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, respErr.ErrorResponse{
			Message: "Internal Server Error",
			Status:  http.StatusInternalServerError,
		})
		return
	}

	ctx.JSON(http.StatusOK, request.SearchResponse{
		Status:  http.StatusOK,
		Data:    dataGI,
		Total:   total,
		Page:    page,
		PerPage: perPage,
	})
}

// Message

func (h *Handler) HandlerSendMessage(ctx *gin.Context) {
	var requestData request.SendMessageRequest
	if err := ctx.ShouldBindJSON(&requestData); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, respErr.ErrorResponse{
			Message: err.Error(),
			Status:  http.StatusBadRequest,
		})
		return
	}

	senderID, _ := ctx.Get("user_id")
	if senderID == nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, respErr.ErrorResponse{
			Message: "User not authenticated",
			Status:  http.StatusUnauthorized,
		})
		return
	}

	senderIDInt64, ok := senderID.(int64)
	if !ok {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, respErr.ErrorResponse{
			Message: "Invalid user_id",
			Status:  http.StatusBadRequest,
		})
		return
	}

	receiverID := requestData.ReceiverID
	messageText := requestData.MessageText

	message := entity.Message{
		SenderID:    senderIDInt64,
		ReceiverID:  receiverID,
		MessageText: messageText,
	}

	createdMessage, err := h.GenshinRepository.CreateMessage(message)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, respErr.ErrorResponse{
			Message: err.Error(),
			Status:  http.StatusInternalServerError,
		})
		return
	}

	ctx.JSON(http.StatusOK, request.ResponseSendMessage{
		Message:        "Message sent successfully",
		CreatedMessage: request.Message(createdMessage),
	})
}

func (h *Handler) HandlerGetMessages(ctx *gin.Context) {
	userID, _ := ctx.Get("user_id")
	if userID == nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, respErr.ErrorResponse{
			Message: "User not authenticated",
			Status:  http.StatusUnauthorized,
		})
		return
	}

	userIDInt64, ok := userID.(int64)
	if !ok {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, respErr.ErrorResponse{
			Message: "Invalid user_id",
			Status:  http.StatusBadRequest,
		})
		return
	}

	messages, err := h.GenshinRepository.GetMessagesByUser(userIDInt64)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, respErr.ErrorResponse{
			Message: err.Error(),
			Status:  http.StatusInternalServerError,
		})
		return
	}

	responseMessages := request.ConvertToResponseMessages(messages)

	ctx.JSON(http.StatusOK, request.ResponseGetMessages{
		Message:  "Messages retrieved successfully",
		Messages: responseMessages,
	})
}

///////////////////////////////////////////////////////////////////

func IsValidEmail(email string) bool {
	// Definisikan pola regex untuk validasi email
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,4}$`
	re := regexp.MustCompile(pattern)

	return re.MatchString(email)
}
