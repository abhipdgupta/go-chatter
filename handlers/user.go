package handlers

import (
	"errors"
	"go-chatter/data"
	"go-chatter/utils"
	"regexp"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

func ValidateUser(user data.User) error {
	if user.Name == "" {
		return errors.New("name is required")
	}

	if len(user.Name) < 2 || len(user.Name) > 50 {
		return errors.New("name must be between 2 and 50 characters")
	}

	if user.Email == "" {
		return errors.New("email is required")
	}

	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(user.Email) {
		return errors.New("invalid email format")
	}

	if user.Password == "" {
		return errors.New("password is required")
	}

	if len(user.Password) < 6 || len(user.Password) > 50 {
		return errors.New("password must be between 6 and 50 characters")
	}

	return nil
}

func RegisterUserRoutes(app *fiber.App) {
	router := app.Group("/user")

	router.Get("/profile", HandleGetUserProfile)
	router.Get("/", HandleGetUser)
	router.Post("/register", HandleRegisterUser)
	router.Post("/login", HandleUserLogin)
}

func HandleUserLogin(c *fiber.Ctx) error {
	// Parse request body into a User struct
	var user data.User
	if err := c.BodyParser(&user); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message":     "Invalid request body",
			"error":       err.Error(),
			"status_code": fiber.StatusBadRequest,
			"data":        nil,
		})
	}

	// Retrieve user from the database by email
	dbUser, err := data.GetUserByEmail(user.Email)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message":     "Failed to retrieve user",
			"error":       err.Error(),
			"status_code": fiber.StatusInternalServerError,
			"data":        nil,
		})
	}

	// Check if user exists
	if dbUser == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message":     "Invalid email or password",
			"status_code": fiber.StatusNotFound,
			"data":        nil,
		})
	}

	// Compare the provided password with the hashed password stored in the database
	err = utils.ComparePassword(dbUser.Password, user.Password)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message":     "Invalid email or password",
			"error":       err.Error(),
			"status_code": fiber.StatusNotFound,
			"data":        nil,
		})
	}

	// Password is correct, generate JWT token
	token, err := utils.SetJwt(dbUser)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message":     "Failed to generate token",
			"error":       err.Error(),
			"status_code": fiber.StatusInternalServerError,
			"data":        nil,
		})
	}

	// Return success response with token
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":     "Login successful",
		"data":        fiber.Map{"token": token},
		"status_code": fiber.StatusOK,
	})
}

func HandleGetUserProfile(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON("ok") // TODO:
}

func HandleGetUser(c *fiber.Ctx) error {
	// Parse query parameters
	query := c.Queries()
	limit, err := strconv.Atoi(query["limit"])
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message":     "Invalid limit query parameter",
			"error":       err.Error(),
			"status_code": fiber.StatusBadRequest,
			"data":        nil,
		})
	}

	page, err := strconv.Atoi(query["page"])
	if err != nil {
		page = 1
	}
	// Retrieve users with pagination
	users, err := data.GetAllUsersWithPagination(page, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message":     "Failed to retrieve users",
			"error":       err.Error(),
			"status_code": fiber.StatusInternalServerError,
			"data":        nil,
		})
	}

	// Return response
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":     "Retrieved users successfully",
		"data":        users,
		"status_code": fiber.StatusOK,
	})
}

func HandleRegisterUser(c *fiber.Ctx) error {
	// Parse request body to extract user information
	var newUser data.User
	if err := c.BodyParser(&newUser); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message":     "Invalid request body",
			"error":       err.Error(),
			"status_code": fiber.StatusBadRequest,
			"data":        nil,
		})
	}

	// Validate user information
	if err := ValidateUser(newUser); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message":     "Validation error",
			"error":       err.Error(),
			"status_code": fiber.StatusBadRequest,
			"data":        nil,
		})
	}

	// Hash the user's password
	hashedPassword, err := utils.HashPassword(newUser.Password)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message":     "Failed to hash password",
			"error":       err.Error(),
			"status_code": fiber.StatusInternalServerError,
			"data":        nil,
		})
	}
	newUser.Password = hashedPassword
	newUser.Role = "USER"
	// Set the creation time
	newUser.CreatedAt = time.Now()

	// Insert the user into the database
	insertedID, err := newUser.InsertUser(newUser)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message":     "Failed to insert user",
			"error":       err.Error(),
			"status_code": fiber.StatusInternalServerError,
			"data":        nil,
		})
	}

	// Return success response
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":     "User inserted successfully",
		"data":        fiber.Map{"inserted_id": insertedID.Hex()},
		"status_code": fiber.StatusCreated,
	})
}
