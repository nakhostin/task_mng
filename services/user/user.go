package user

import (
	"errors"
	"fmt"
	"log/slog"
	"task_mng/domain/user"
	"task_mng/domain/user/aggregate"
	"task_mng/domain/user/entity"
	"task_mng/pkg/jwt"

	"github.com/asaskevich/govalidator"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Service struct {
	repository user.Repository
	logger     *slog.Logger
	jwtManager jwt.JWTManager
}

func New(repository user.Repository, jwtManager jwt.JWTManager) *Service {
	return &Service{repository: repository, logger: slog.Default(), jwtManager: jwtManager}
}

// ********************* Create *********************
type CreateRequest struct {
	Username string `json:"username" valid:"required~username_is_required,length(3|20)~username_must_be_3_to_20_characters" example:"admin"`
	FullName string `json:"full_name" valid:"required~full_name_is_required" example:"Admin"`
	Email    string `json:"email" valid:"required~email_is_required,email~email_is_invalid" example:"admin@xdr.com"`
	Password string `json:"password" valid:"required~password_is_required,length(8|32)~password_must_be_8_to_32_characters" example:"Admin!123"`
}

func (s *Service) Create(req *CreateRequest) error {
	user, err := s.repository.FindByUsername(req.Username)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Error("error finding user", "error", err)
			return err
		}
	}

	if user.ID != 0 {
		s.logger.Error("user already exists")
		return errors.New("user_already_exists")
	}

	hashedPassword, err := s.hashPassword(req.Password)
	if err != nil {
		s.logger.Error("error hashing password", "error", err)
		return errors.New("internal_server_error")
	}

	err = s.repository.Create(&entity.User{
		Username: req.Username,
		FullName: req.FullName,
		Email:    req.Email,
		Password: hashedPassword,
	})
	if err != nil {
		s.logger.Error("error creating user", "error", err)
		return errors.New("internal_server_error")
	}

	return nil
}

// ********************* Login *********************
type LoginRequest struct {
	Username string `json:"username" valid:"required~username_is_required,length(3|20)~username_must_be_3_to_20_characters" example:"admin"`
	Password string `json:"password" valid:"required~password_is_required,length(8|32)~password_must_be_8_to_32_characters" example:"Admin!123"`
}

func (s *Service) Login(req *LoginRequest) (*aggregate.AuthResponse, error) {
	user, err := s.repository.FindByUsername(req.Username)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Error("error finding user", "error", err)
			return nil, errors.New("username_or_password_is_incorrect")
		}
	}

	if !s.verifyPassword(user.Password, req.Password) {
		s.logger.Error("password is incorrect")
		return nil, errors.New("username_or_password_is_incorrect")
	}

	tokens, err := s.jwtManager.GenerateTokenPair(fmt.Sprint(user.ID), user.Email, user.Username, "")
	if err != nil {
		s.logger.Error("error generating token pair", "error", err)
		return nil, errors.New("internal_server_error")
	}

	return &aggregate.AuthResponse{
		User:   aggregate.NewUserResponse(&user),
		Tokens: tokens,
	}, nil
}

// ********************* Refresh Token *********************
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" valid:"required~refresh_token_is_required" example:"refresh_token"`
}

func (s *Service) Refresh(req *RefreshRequest) (*aggregate.AuthResponse, error) {
	tokens, err := s.jwtManager.GenerateNewTokenPair(req.RefreshToken)
	if err != nil {
		s.logger.Error("error generating new token pair", "error", err)
		return nil, err
	}

	return &aggregate.AuthResponse{
		Tokens: tokens,
	}, nil
}

// ********************* Find By ID *********************
func (s *Service) FindByID(id uint) (*aggregate.UserResponse, error) {
	usr, err := s.repository.FindByID(id)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Error("error finding user", "error", err)
			return nil, err
		}
		s.logger.Error("error finding user", "error", err)
		return nil, err
	}
	return aggregate.NewUserResponse(&usr), nil
}

// ********************* Update Profile *********************
type UpdateProfileRequest struct {
	FullName string `json:"full_name" valid:"required~full_name_is_required" example:"Admin"`
	Email    string `json:"email" valid:"required~email_is_required,email~email_is_invalid" example:"admin@xdr.com"`
}

func (s *Service) UpdateProfile(id uint, req *UpdateProfileRequest) (*aggregate.UserResponse, error) {
	if _, err := govalidator.ValidateStruct(req); err != nil {
		s.logger.Error("error validating request", "error", err)
		return nil, err
	}

	usr, err := s.repository.FindByID(id)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Error("error finding user", "error", err)
			return nil, fmt.Errorf("internal_server_error")
		}
	}

	if usr.ID == 0 {
		s.logger.Error("user not found")
		return nil, fmt.Errorf("user_not_found")
	}

	usr.FullName = req.FullName
	usr.Email = req.Email
	err = s.repository.Update(usr)
	if err != nil {
		s.logger.Error("error updating user", "error", err)
		return nil, fmt.Errorf("internal_server_error")
	}

	return aggregate.NewUserResponse(&usr), nil
}

// ********************* Find All *********************
func (s *Service) FindAll(page, limit int) (*aggregate.UserListResponse, error) {
	users, count, err := s.repository.FindAll(page, limit)
	if err != nil {
		s.logger.Error("error finding users", "error", err)
		return nil, err
	}

	return aggregate.NewUserListResponse(users, page, limit, count, ""), nil
}

// Helper functions
func (s *Service) hashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("error hashing password", "error", err)
		return "", err
	}
	return string(hashedPassword), nil
}

func (s *Service) verifyPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		s.logger.Error("error verifying password", "error", err)
		return false
	}
	return true
}
