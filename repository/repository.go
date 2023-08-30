package repository

import (
	"ListCharacterGI/model/entity"
	"io"
	"mime/multipart"
)

type GenshinRepository interface {
	GetAllUsers() ([]entity.User, error)
	DeleteUserByIDAndRole(userID int64, role string) error

	// User
	CreateUser(user *entity.User) error

	// ALL //////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	GetAllCharacters() ([]entity.Characters, error)
	GetByID(Id int64) (*entity.Characters, error)
	Create(mahasiswa *entity.Characters) (*entity.Characters, error)
	GetCharacterByName(name string) (*entity.Characters, error)
	Update(character *entity.Characters) (*entity.Characters, error)
	UpdatetoAtch(todo *entity.Characters) error
	Delete(Id int64) (int64, error)
	GetUserByUsernameOrEmail(username, email string) (*entity.User, error)
	//UploadTodoFileS3(file *multipart.FileHeader, url string) error
	//UploadTodoFileLocal(file *multipart.FileHeader, url string) error
	/////////////////////
	CreateAttachment(mhsID int64, path string, order int64) (*entity.Attachment, error)
	UploadFileS3Atch(file *multipart.FileHeader, Id int64) (*entity.Attachment, error)
	UpdateWithAttachments(mhs *entity.Characters) error
	UploadFileS3Buckets(file io.Reader, fileName string) (*string, error)
	UploadFileLocalAtch(file *multipart.FileHeader, id int64) (*entity.Attachment, error)
	GeneratePresignedURL(bucketName, objectKey string) (string, error)
	GetAttachmentByID(attachmentID int64) (*entity.Attachment, error)
	//
	SearchCharacterByUser(search string, page, perPage int) ([]entity.Characters, int64, error)
	//

	// Chat antar pengguna
	CreateMessage(message entity.Message) (entity.Message, error)
	GetMessagesByUser(userID int64) ([]entity.Message, error)
}
