package database

import (
	"ListCharacterGI/model/entity"
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"
)

// adaptop pattern
type GenshinRepository struct {
	DB       *gorm.DB
	S3Bucket *s3.Client
}

func NewGenshinRepository(DB *gorm.DB, s3Bucket *s3.Client) *GenshinRepository {
	return &GenshinRepository{
		DB:       DB,
		S3Bucket: s3Bucket,
	}
}

//func (t GenshinRepository) GetAll() ([]entity.Mahasiswa, error) {
//	var data []entity.Mahasiswa
//
//	result := t.DB.Preload("Attachments").Preload("User").Find(&data)
//	return data, result.Error
//}

func (t *GenshinRepository) GetAllUsers() ([]entity.User, error) {
	var users []entity.User
	if err := t.DB.Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (t *GenshinRepository) DeleteUserByIDAndRole(userID int64, role string) error {
	// Cek apakah pengguna dengan ID dan peran tertentu ada dalam database
	user := &entity.User{}
	err := t.DB.Where("id = ? AND role = ?", userID, role).First(user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("user not found or does not have the specified role")
		}
		return err
	}

	// Hapus pengguna dari database
	if err := t.DB.Delete(user).Error; err != nil {
		return err
	}

	return nil
}

func (t GenshinRepository) GetAllCharacters() ([]entity.Characters, error) {
	var data []entity.Characters

	// Ambil semua data berdasarkan user_id
	result := t.DB.Preload("Attachments").Find(&data)
	if result.Error != nil {
		return nil, result.Error
	}

	return data, nil
}

func (t GenshinRepository) GetByID(Id int64) (*entity.Characters, error) {
	var data entity.Characters
	result := t.DB.Preload("Attachments").Where("id = ?", Id).First(&data)
	if result.Error != nil {
		return nil, result.Error
	}

	return &data, nil
}

func (t GenshinRepository) Create(mahasiswa *entity.Characters) (*entity.Characters, error) {
	result := t.DB.Create(mahasiswa)
	return mahasiswa, result.Error

}

func (t GenshinRepository) GetCharacterByName(name string) (*entity.Characters, error) {
	var data entity.Characters
	result := t.DB.Where("name = ?", name).First(&data)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &data, nil
}

func (t GenshinRepository) Update(character *entity.Characters) (*entity.Characters, error) {
	result := t.DB.Model(&entity.Characters{}).Where("id = ?", character.ID).Updates(character)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, result.Error
	}
	return character, nil
}

func (t GenshinRepository) UpdatetoAtch(mhs *entity.Characters) error {
	err := t.DB.Save(mhs).Error
	return err
}

func (t GenshinRepository) Delete(Id int64) (int64, error) {
	data := entity.Characters{}

	// Fetch the data by ID and user_id
	if err := t.DB.Where("id = ?", Id).First(&data).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// If data not found, return 0 RowsAffected
			return 0, err
		}
		return 0, err
	}

	// Delete the fetched data
	result := t.DB.Delete(&data)
	return result.RowsAffected, nil
}

func (t GenshinRepository) CreateUser(user *entity.User) error {
	if err := t.DB.Create(user).Error; err != nil {
		return err
	}
	return nil
}

func (t GenshinRepository) GetUserByUsernameOrEmail(username, email string) (*entity.User, error) {
	var user entity.User
	result := t.DB.Where("username = ? OR email = ?", username, email).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &user, nil

}

/////////////////////////////////////////

func (t *GenshinRepository) CreateAttachment(mhsID int64, path string, order int64) (*entity.Attachment, error) {
	attachment := &entity.Attachment{
		UserID:          mhsID,
		Path:            path,
		AttachmentOrder: order,
	}
	if err := t.DB.Create(attachment).Error; err != nil {
		return nil, err
	}
	return attachment, nil
}

func (t *GenshinRepository) UploadFileS3Atch(file *multipart.FileHeader, Id int64) (*entity.Attachment, error) {
	//Mengambil data berdasarkan ID dan user_id

	dataMhs := &entity.Characters{}
	if err := t.DB.Where("id = ?", Id).First(dataMhs).Error; err != nil {
		return nil, err
	}

	src, err := file.Open()
	if err != nil {
		logrus.Error(err)
		return nil, err
	}
	defer src.Close()

	// bikin nama file yang uniq untuk menghindari konflik
	uniqueFilename := fmt.Sprintf("%s%s", uuid.NewString(), filepath.Ext(file.Filename))

	// Upload the file to S3
	bucketName := "bucketwithrey"
	objectKey := uniqueFilename
	_, err = t.S3Bucket.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
		Body:   src,
		//ACL:    types.ObjectCannedACLPublicRead, // Optional: Mengatur ACL agar file yang diunggah dapat diakses oleh publik
	})
	if err != nil {
		logrus.Error(err)
		return nil, err
	}

	// Return the public URL of the uploaded file
	//publicURL := fmt.Sprintf("https://%s.s3.amazonaws.com/%s", bucketName, objectKey)

	// Create an attachment record in the database
	var attachmentOrder int64 = 1 // Set the initial attachment_order
	// Get the count of existing attachments for the dataMhs
	existingAttachmentCount := int64(0)
	attachmentOrder = existingAttachmentCount + 1 // Set attachment_order dynamically

	// Create an attachment record in the database
	attachment := &entity.Attachment{
		UserID:          Id,
		Path:            objectKey,
		AttachmentOrder: attachmentOrder, // atur order
		Timestamp:       time.Now(),
	}
	err = t.DB.Create(attachment).Error
	if err != nil {
		return nil, err
	}

	return attachment, nil
}
func (t *GenshinRepository) UpdateWithAttachments(mhs *entity.Characters) error {
	return t.DB.Transaction(func(tx *gorm.DB) error {
		// Pertama, hapus semua lampiran yang ada yang terkait dengan data
		if err := t.DB.Where("user_id = ?", mhs.ID).Delete(&entity.Attachment{}).Error; err != nil {
			return err
		}

		// Next, create new attachment records
		for i := range mhs.Attachments {
			attachment := &mhs.Attachments[i]
			if err := t.DB.Create(attachment).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (t *GenshinRepository) UploadFileS3Buckets(file io.Reader, fileName string) (*string, error) {
	bucketName := "bucketwithrey"
	objectKey := fileName

	_, err := t.S3Bucket.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
		Body:   file,
	})
	if err != nil {
		logrus.Error(err)
		return nil, err
	}

	// Return the public URL of the uploaded file
	publicURL := aws.String(fmt.Sprintf("https://%s.s3.amazonaws.com/%s", bucketName, objectKey))

	return publicURL, nil
}

func (t *GenshinRepository) UploadFileLocalAtch(file *multipart.FileHeader, Id int64) (*entity.Attachment, error) {
	// Fetch the data by ID and user_id
	data := &entity.Characters{}
	if err := t.DB.Where("id = ?", Id).First(data).Error; err != nil {
		return nil, err
	}

	src, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer src.Close()

	// bikin nama file yang uniq untuk menghindari konflik
	uniqueFilename := fmt.Sprintf("%s%s", uuid.NewString(), filepath.Ext(file.Filename))

	// Upload the file to Local
	uploadDir := "uploads"

	// Buat direktori unggahan jika belum ada
	err = os.MkdirAll(uploadDir, 0755)
	if err != nil {
		return nil, err
	}

	// Create the destination file
	dest, err := os.Create(filepath.Join(uploadDir, uniqueFilename))
	if err != nil {
		return nil, err
	}
	defer dest.Close()

	// Copy file nya ke file tujuan
	_, err = io.Copy(dest, src)
	if err != nil {
		return nil, err
	}
	// Return the local file path
	localFilePath := filepath.Join(uploadDir, uniqueFilename)

	// Create an attachment record in the database
	var attachmentOrder int64 = 1 // Set the initial attachment_order
	// Get the count of existing attachments for the dataMhs
	existingAttachmentCount := int64(0)
	attachmentOrder = existingAttachmentCount + 1 // Set attachment_order dynamically

	// Create an attachment record in the database
	attachment := &entity.Attachment{
		UserID:          Id,
		Path:            localFilePath,
		AttachmentOrder: attachmentOrder, // atur order
		Timestamp:       time.Now(),
	}
	err = t.DB.Create(attachment).Error
	if err != nil {
		return nil, err
	}

	return attachment, nil
}

func (t *GenshinRepository) GeneratePresignedURL(bucketName, objectKey string) (string, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return "", err
	}

	s3client := s3.NewFromConfig(cfg)

	presignClient := s3.NewPresignClient(s3client)
	req, err := presignClient.PresignGetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
	},
		s3.WithPresignExpires(time.Minute*15),
	)

	return req.URL, nil
}

func (t *GenshinRepository) GetAttachmentByID(attachmentID int64) (*entity.Attachment, error) {
	var attachment entity.Attachment
	if err := t.DB.First(&attachment, attachmentID).Error; err != nil {
		return nil, err
	}
	return &attachment, nil
}

func (t *GenshinRepository) SearchCharacterByUser(search string, page, perPage int) ([]entity.Characters, int64, error) {
	var dataGI []entity.Characters

	// Menghitung total data
	var total int64
	t.DB.Model(&entity.Characters{}).
		Where("name LIKE ? OR address LIKE ? OR element = ? OR weapon_type = ? OR star_rating = ?",
			"%"+search+"%", "%"+search+"%", search, search, search).
		Count(&total)

	// Mengambil data dengan paginasi
	offset := (page - 1) * perPage
	err := t.DB.Where("name LIKE ? OR address LIKE ? OR element = ? OR weapon_type = ? OR star_rating = ?",
		"%"+search+"%", "%"+search+"%", search, search, search).
		Offset(offset).Limit(perPage).
		Preload("Attachments").Find(&dataGI).Error

	return dataGI, total, err
}

// Messages
func (t GenshinRepository) CreateMessage(message entity.Message) (entity.Message, error) {
	result := t.DB.Create(&message)
	if result.Error != nil {
		return entity.Message{}, result.Error
	}
	return message, nil
}

func (t GenshinRepository) GetMessagesByUser(userID int64) ([]entity.Message, error) {
	var messages []entity.Message
	result := t.DB.Where("sender_id = ? OR receiver_id = ?", userID, userID).Find(&messages)
	if result.Error != nil {
		return nil, result.Error
	}
	return messages, nil
}
