package handlers

import (
	"task_mng/pkg/response"
	"task_mng/services/user"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService *user.Service
}

func NewUserHandler(userService *user.Service) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

type CreateRequest struct {
	Username string `json:"username" valid:"required~username_is_required,length(3|32)~username_must_be_3_to_32_characters"`
	FullName string `json:"full_name" valid:"required~full_name_is_required"`
	Email    string `json:"email" valid:"required~email_is_required,email~email_is_invalid"`
	Password string `json:"password" valid:"required~password_is_required,length(8|32)~password_must_be_8_to_32_characters"`
}

// Create godoc
// @Summary Create a new user
// @Description Create a new user in the system (requires authentication)
// @Tags Users
// @Accept json
// @Produce json
// @Param request body user.CreateRequest true "User registration data"
// @Success 200 {object} response.Response "Create successful"
// @Failure 400 {object} response.Response "Bad request"
// @Security BearerAuth
// @Router /users [post]
func (h *UserHandler) Create(c *gin.Context) {
	req, err := response.Parse[user.CreateRequest](c)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	err = h.userService.Create(req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, "Create successful", nil, nil)
}

type LoginRequest struct {
	Username string `json:"username" valid:"required~username_is_required,length(3|32)~username_must_be_3_to_32_characters"`
	Password string `json:"password" valid:"required~password_is_required,length(8|32)~password_must_be_8_to_32_characters"`
}

// Login godoc
// @Summary User login
// @Description Authenticate user and return tokens
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body user.LoginRequest true "User login credentials"
// @Success 200 {object} response.Response{data=aggregate.AuthResponse} "Login successful"
// @Failure 400 {object} response.Response "Bad request"
// @Router /auth/login [post]
func (h *UserHandler) Login(c *gin.Context) {
	req, err := response.Parse[user.LoginRequest](c)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	resp, err := h.userService.Login(req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, "Login successful", resp, nil)
}

// Refresh godoc
// @Summary Refresh access token
// @Description Get a new access token using refresh token
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body user.RefreshRequest true "Refresh token"
// @Success 200 {object} response.Response{data=aggregate.AuthResponse} "Tokens refreshed successful"
// @Failure 400 {object} response.Response "Bad request"
// @Router /auth/refresh [post]
func (h *UserHandler) Refresh(c *gin.Context) {
	req, err := response.Parse[user.RefreshRequest](c)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	resp, err := h.userService.Refresh(req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, "Tokens refreshed successful", resp, nil)
}

// Me godoc
// @Summary Get current user profile
// @Description Get the authenticated user's profile information
// @Tags Profile
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=aggregate.UserResponse} "User fetched"
// @Failure 400 {object} response.Response "Bad request"
// @Security BearerAuth
// @Router /profile [get]
func (h *UserHandler) Me(c *gin.Context) {
	userID, _ := c.Get("user_id")
	user, err := h.userService.FindByID(userID.(uint))
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, "User fetched", user, nil)
}

// Update godoc
// @Summary Update user profile
// @Description Update the authenticated user's profile information
// @Tags Profile
// @Accept json
// @Produce json
// @Param request body user.UpdateProfileRequest true "Profile update data"
// @Success 200 {object} response.Response{data=aggregate.UserResponse} "Profile updated successfully"
// @Failure 400 {object} response.Response "Bad request"
// @Security BearerAuth
// @Router /profile [put]
func (h *UserHandler) Update(c *gin.Context) {
	userID, _ := c.Get("user_id")

	req, err := response.Parse[user.UpdateProfileRequest](c)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	resp, err := h.userService.UpdateProfile(userID.(uint), req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, "Profile updated successfully", resp, nil)
}

// FindAll godoc
// @Summary Get all users
// @Description Get a list of all users with pagination
// @Tags Users
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} response.Response{data=aggregate.UserListResponse} "Users fetched successfully"
// @Failure 400 {object} response.Response "Bad request"
// @Security BearerAuth
// @Router /users [get]
func (h *UserHandler) FindAll(c *gin.Context) {
	pag := response.NewPagination(c)

	result, err := h.userService.FindAll(pag.Page, pag.Limit)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, "Users fetched successfully", result, result.Meta)
}
