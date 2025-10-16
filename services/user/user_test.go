package user

import (
	"errors"
	"task_mng/domain/user/entity"
	"task_mng/domain/user/mocks"
	jwtMocks "task_mng/pkg/jwt/mocks"
	"testing"
	"time"

	"task_mng/pkg/jwt"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// ********************* Register Tests *********************

func TestCreate_Success(t *testing.T) {
	mockRepo := new(mocks.MockUserRepository)
	mockJWT := &jwtMocks.MockJWTManager{}
	service := New(mockRepo, mockJWT)

	req := &CreateRequest{
		Username: "user",
		FullName: "User",
		Email:    "user@example.com",
		Password: "Password!123",
	}

	// Mock: user not found (new user)
	mockRepo.On("FindByUsername", req.Username).Return(entity.User{}, gorm.ErrRecordNotFound)

	// Mock: create user
	mockRepo.On("Create", mock.MatchedBy(func(u *entity.User) bool {
		return u.Username == req.Username &&
			u.FullName == req.FullName &&
			u.Email == req.Email &&
			u.Password != "" // password should be hashed
	})).Return(nil)

	err := service.Create(req)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestCreate_UserAlreadyExists(t *testing.T) {
	mockRepo := new(mocks.MockUserRepository)
	mockJWT := &jwtMocks.MockJWTManager{}
	service := New(mockRepo, mockJWT)

	req := &CreateRequest{
		Username: "user",
		FullName: "User",
		Email:    "user@example.com",
		Password: "Password!123",
	}

	existingUser := entity.User{
		Model:    gorm.Model{ID: 1},
		Username: req.Username,
	}

	mockRepo.On("FindByUsername", req.Username).Return(existingUser, nil)

	err := service.Create(req)

	assert.Error(t, err)
	assert.Equal(t, "user_already_exists", err.Error())
	mockRepo.AssertExpectations(t)
}

func TestCreate_RepositoryFindError(t *testing.T) {
	mockRepo := new(mocks.MockUserRepository)
	mockJWT := &jwtMocks.MockJWTManager{}
	service := New(mockRepo, mockJWT)

	req := &CreateRequest{
		Username: "user",
		FullName: "User",
		Email:    "user@example.com",
		Password: "Password!123",
	}

	dbError := errors.New("database error")
	mockRepo.On("FindByUsername", req.Username).Return(entity.User{}, dbError)

	err := service.Create(req)

	assert.Error(t, err)
	assert.Equal(t, dbError, err)
	mockRepo.AssertExpectations(t)
}

func TestCreate_CreateError(t *testing.T) {
	mockRepo := new(mocks.MockUserRepository)
	mockJWT := &jwtMocks.MockJWTManager{}
	service := New(mockRepo, mockJWT)

	req := &CreateRequest{
		Username: "user",
		FullName: "User",
		Email:    "user@example.com",
		Password: "Password!123",
	}

	mockRepo.On("FindByUsername", req.Username).Return(entity.User{}, gorm.ErrRecordNotFound)

	createError := errors.New("create error")
	mockRepo.On("Create", mock.Anything).Return(createError)

	err := service.Create(req)

	assert.Error(t, err)
	assert.Equal(t, "internal_server_error", err.Error())
	mockRepo.AssertExpectations(t)
}

// ********************* Login Tests *********************

func TestLogin_Success(t *testing.T) {
	mockRepo := new(mocks.MockUserRepository)
	mockJWT := &jwtMocks.MockJWTManager{}
	service := New(mockRepo, mockJWT)

	password := "Password!123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	existingUser := entity.User{
		Model: gorm.Model{
			ID:        1,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Username: "user",
		FullName: "User",
		Email:    "user@example.com",
		Password: string(hashedPassword),
	}

	req := &LoginRequest{
		Username: "user",
		Password: password,
	}

	expectedTokens := &jwt.TokenPair{
		AccessToken:  "access_token",
		RefreshToken: "refresh_token",
		ExpiresAt:    time.Now().Add(15 * time.Minute),
		TokenType:    "Bearer",
	}

	mockRepo.On("FindByUsername", req.Username).Return(existingUser, nil)
	mockJWT.GenerateTokenPairFunc = func(userID, email, username, role string) (*jwt.TokenPair, error) {
		assert.Equal(t, "1", userID)
		assert.Equal(t, existingUser.Email, email)
		assert.Equal(t, existingUser.Username, username)
		return expectedTokens, nil
	}

	result, err := service.Login(req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, result.User)
	assert.NotNil(t, result.Tokens)
	assert.Equal(t, uint(1), result.User.ID)
	assert.Equal(t, existingUser.Username, result.User.Username)
	assert.Equal(t, expectedTokens.AccessToken, result.Tokens.AccessToken)
	mockRepo.AssertExpectations(t)
}

func TestLogin_UserNotFound(t *testing.T) {
	mockRepo := new(mocks.MockUserRepository)
	mockJWT := &jwtMocks.MockJWTManager{}
	service := New(mockRepo, mockJWT)

	req := &LoginRequest{
		Username: "user",
		Password: "Password!123",
	}

	mockRepo.On("FindByUsername", req.Username).Return(entity.User{}, gorm.ErrRecordNotFound)

	result, err := service.Login(req)

	assert.Error(t, err)
	assert.Nil(t, result)
	// Note: Login returns empty user when not found, which fails password verification
	mockRepo.AssertExpectations(t)
}

func TestLogin_InvalidPassword(t *testing.T) {
	mockRepo := new(mocks.MockUserRepository)
	mockJWT := &jwtMocks.MockJWTManager{}
	service := New(mockRepo, mockJWT)

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correct_password"), bcrypt.DefaultCost)

	existingUser := entity.User{
		Model:    gorm.Model{ID: 1},
		Username: "user",
		Password: string(hashedPassword),
	}

	req := &LoginRequest{
		Username: "user",
		Password: "wrong_password",
	}

	mockRepo.On("FindByUsername", req.Username).Return(existingUser, nil)

	result, err := service.Login(req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "username_or_password_is_incorrect", err.Error())
	mockRepo.AssertExpectations(t)
}

func TestLogin_TokenGenerationError(t *testing.T) {
	mockRepo := new(mocks.MockUserRepository)
	mockJWT := &jwtMocks.MockJWTManager{}
	service := New(mockRepo, mockJWT)

	password := "Password!123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	existingUser := entity.User{
		Model:    gorm.Model{ID: 1},
		Username: "user",
		Email:    "nima@example.com",
		Password: string(hashedPassword),
	}

	req := &LoginRequest{
		Username: "user",
		Password: password,
	}

	mockRepo.On("FindByUsername", req.Username).Return(existingUser, nil)
	mockJWT.GenerateTokenPairFunc = func(userID, email, username, role string) (*jwt.TokenPair, error) {
		return nil, errors.New("token generation error")
	}

	result, err := service.Login(req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "internal_server_error", err.Error())
	mockRepo.AssertExpectations(t)
}

func TestLogin_RepositoryError(t *testing.T) {
	mockRepo := new(mocks.MockUserRepository)
	mockJWT := &jwtMocks.MockJWTManager{}
	service := New(mockRepo, mockJWT)

	req := &LoginRequest{
		Username: "user",
		Password: "Password!123",
	}

	dbError := errors.New("database error")
	mockRepo.On("FindByUsername", req.Username).Return(entity.User{}, dbError)

	result, err := service.Login(req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "username_or_password_is_incorrect", err.Error())
	mockRepo.AssertExpectations(t)
}

// ********************* Refresh Token Tests *********************

func TestRefresh_Success(t *testing.T) {
	mockRepo := new(mocks.MockUserRepository)
	mockJWT := &jwtMocks.MockJWTManager{}
	service := New(mockRepo, mockJWT)

	req := &RefreshRequest{
		RefreshToken: "valid_refresh_token",
	}

	expectedTokens := &jwt.TokenPair{
		AccessToken:  "new_access_token",
		RefreshToken: "",
		ExpiresAt:    time.Now().Add(15 * time.Minute),
		TokenType:    "Bearer",
	}

	mockJWT.GenerateNewTokenPairFunc = func(refreshToken string) (*jwt.TokenPair, error) {
		assert.Equal(t, req.RefreshToken, refreshToken)
		return expectedTokens, nil
	}

	result, err := service.Refresh(req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, result.Tokens)
	assert.Equal(t, expectedTokens.AccessToken, result.Tokens.AccessToken)
}

func TestRefresh_InvalidToken(t *testing.T) {
	mockRepo := new(mocks.MockUserRepository)
	mockJWT := &jwtMocks.MockJWTManager{}
	service := New(mockRepo, mockJWT)

	req := &RefreshRequest{
		RefreshToken: "invalid_refresh_token",
	}

	mockJWT.GenerateNewTokenPairFunc = func(refreshToken string) (*jwt.TokenPair, error) {
		return nil, jwt.ErrInvalidToken
	}

	result, err := service.Refresh(req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, jwt.ErrInvalidToken, err)
}

func TestRefresh_ExpiredToken(t *testing.T) {
	mockRepo := new(mocks.MockUserRepository)
	mockJWT := &jwtMocks.MockJWTManager{}
	service := New(mockRepo, mockJWT)

	req := &RefreshRequest{
		RefreshToken: "expired_refresh_token",
	}

	mockJWT.GenerateNewTokenPairFunc = func(refreshToken string) (*jwt.TokenPair, error) {
		return nil, jwt.ErrExpiredToken
	}

	result, err := service.Refresh(req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, jwt.ErrExpiredToken, err)
}

// ********************* FindByID Tests *********************

func TestFindByID_Success(t *testing.T) {
	mockRepo := new(mocks.MockUserRepository)
	mockJWT := &jwtMocks.MockJWTManager{}
	service := New(mockRepo, mockJWT)

	expectedUser := entity.User{
		Model: gorm.Model{
			ID:        1,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Username: "user",
		FullName: "User",
		Email:    "user@example.com",
	}

	mockRepo.On("FindByID", uint(1)).Return(expectedUser, nil)

	result, err := service.FindByID(1)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, uint(1), result.ID)
	assert.Equal(t, expectedUser.Username, result.Username)
	assert.Equal(t, expectedUser.FullName, result.FullName)
	assert.Equal(t, expectedUser.Email, result.Email)

	mockRepo.AssertExpectations(t)
}

func TestFindByID_NotFound(t *testing.T) {
	mockRepo := new(mocks.MockUserRepository)
	mockJWT := &jwtMocks.MockJWTManager{}
	service := New(mockRepo, mockJWT)

	mockRepo.On("FindByID", uint(999)).Return(entity.User{}, gorm.ErrRecordNotFound)

	result, err := service.FindByID(999)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, gorm.ErrRecordNotFound, err)

	mockRepo.AssertExpectations(t)
}

func TestFindByID_RepositoryError(t *testing.T) {
	mockRepo := new(mocks.MockUserRepository)
	mockJWT := &jwtMocks.MockJWTManager{}
	service := New(mockRepo, mockJWT)

	dbError := errors.New("database error")
	mockRepo.On("FindByID", uint(1)).Return(entity.User{}, dbError)

	result, err := service.FindByID(1)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, dbError, err)

	mockRepo.AssertExpectations(t)
}

// ********************* UpdateProfile Tests *********************

func TestUpdateProfile_Success(t *testing.T) {
	mockRepo := new(mocks.MockUserRepository)
	mockJWT := &jwtMocks.MockJWTManager{}
	service := New(mockRepo, mockJWT)

	existingUser := entity.User{
		Model: gorm.Model{
			ID:        1,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Username: "user",
		FullName: "User",
		Email:    "user@example.com",
	}

	req := &UpdateProfileRequest{
		FullName: "New User",
		Email:    "newuser@example.com",
	}

	mockRepo.On("FindByID", uint(1)).Return(existingUser, nil)
	mockRepo.On("Update", mock.MatchedBy(func(u entity.User) bool {
		return u.ID == 1 &&
			u.FullName == req.FullName &&
			u.Email == req.Email
	})).Return(nil)

	result, err := service.UpdateProfile(1, req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, uint(1), result.ID)
	assert.Equal(t, req.FullName, result.FullName)
	assert.Equal(t, req.Email, result.Email)

	mockRepo.AssertExpectations(t)
}

func TestUpdateProfile_ValidationError(t *testing.T) {
	mockRepo := new(mocks.MockUserRepository)
	mockJWT := &jwtMocks.MockJWTManager{}
	service := New(mockRepo, mockJWT)

	req := &UpdateProfileRequest{
		FullName: "User",
		Email:    "invalid-email",
	}

	result, err := service.UpdateProfile(1, req)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestUpdateProfile_UserNotFound(t *testing.T) {
	mockRepo := new(mocks.MockUserRepository)
	mockJWT := &jwtMocks.MockJWTManager{}
	service := New(mockRepo, mockJWT)

	req := &UpdateProfileRequest{
		FullName: "New Name",
		Email:    "new@example.com",
	}

	mockRepo.On("FindByID", uint(999)).Return(entity.User{}, gorm.ErrRecordNotFound)

	result, err := service.UpdateProfile(999, req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "user_not_found", err.Error())

	mockRepo.AssertExpectations(t)
}

func TestUpdateProfile_UpdateError(t *testing.T) {
	mockRepo := new(mocks.MockUserRepository)
	mockJWT := &jwtMocks.MockJWTManager{}
	service := New(mockRepo, mockJWT)

	existingUser := entity.User{
		Model:    gorm.Model{ID: 1},
		Username: "user",
		FullName: "User",
		Email:    "user@example.com",
	}

	req := &UpdateProfileRequest{
		FullName: "New User",
		Email:    "newuser@example.com",
	}

	updateError := errors.New("update error")
	mockRepo.On("FindByID", uint(1)).Return(existingUser, nil)
	mockRepo.On("Update", mock.Anything).Return(updateError)

	result, err := service.UpdateProfile(1, req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "internal_server_error", err.Error())

	mockRepo.AssertExpectations(t)
}

func TestUpdateProfile_FindByIDError(t *testing.T) {
	mockRepo := new(mocks.MockUserRepository)
	mockJWT := &jwtMocks.MockJWTManager{}
	service := New(mockRepo, mockJWT)

	req := &UpdateProfileRequest{
		FullName: "New User",
		Email:    "newuser@example.com",
	}

	dbError := errors.New("database error")
	mockRepo.On("FindByID", uint(1)).Return(entity.User{}, dbError)

	result, err := service.UpdateProfile(1, req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "internal_server_error", err.Error())

	mockRepo.AssertExpectations(t)
}

// ********************* Helper Function Tests *********************

func TestHashPassword(t *testing.T) {
	mockRepo := new(mocks.MockUserRepository)
	mockJWT := &jwtMocks.MockJWTManager{}
	service := New(mockRepo, mockJWT)

	password := "Password!123"
	hashedPassword, err := service.hashPassword(password)

	assert.NoError(t, err)
	assert.NotEmpty(t, hashedPassword)
	assert.NotEqual(t, password, hashedPassword)

	// Verify the hashed password can be verified
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	assert.NoError(t, err)
}

func TestVerifyPassword_Success(t *testing.T) {
	mockRepo := new(mocks.MockUserRepository)
	mockJWT := &jwtMocks.MockJWTManager{}
	service := New(mockRepo, mockJWT)

	password := "Password!123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	result := service.verifyPassword(string(hashedPassword), password)
	assert.True(t, result)
}

func TestVerifyPassword_Failure(t *testing.T) {
	mockRepo := new(mocks.MockUserRepository)
	mockJWT := &jwtMocks.MockJWTManager{}
	service := New(mockRepo, mockJWT)

	password := "Password!123"
	wrongPassword := "wrong_password"
	result := service.verifyPassword(password, wrongPassword)

	assert.False(t, result)
}
