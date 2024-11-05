package user

import (
	"fmt"
	"log"
	"net/http"

	"RefCode.com/m/internal/storage"
	"RefCode.com/m/types"

	"RefCode.com/m/utils"
	"github.com/go-playground/validator/v10"
)

type Handler struct {
	store types.UserStore
}

func NewHandler(store types.UserStore) *Handler {
	return &Handler{store: store}
}

func (h Handler) HandlerLogin(w http.ResponseWriter, r *http.Request) {
	var payload types.LoginUserPayload
	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}
	// Validations
	if err := utils.Validate.Struct(payload); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload: %v", validationErrors))
		return
	}

	// Получаем пользователя по email
	user, err := h.store.GetUserByEmail(payload.Email)
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, fmt.Errorf("user not found: %w", err))
		return
	}

	if !utils.CheckPasswordHash(payload.Password, user.Password) {
		utils.WriteError(w, http.StatusUnauthorized, fmt.Errorf("invalid credentials"))
		return
	}

	// Генерация JWT-токена
	token, err := utils.GenerateJWT(user.ID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("could not generate token: %w", err))
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{"token": token})
}

func (h Handler) HandlerRegister(w http.ResponseWriter, r *http.Request) {
	// Utils parse JSON
	var payload types.RegisterUserPayload
	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}
	// Utils Validations
	if err := utils.Validate.Struct(payload); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload: %v", validationErrors))
		return
	}
	// Check user
	_, err := h.store.GetUserByEmail(payload.Email)
	if err == nil {
		utils.WriteError(w, http.StatusBadRequest, storage.ErrDuplicateEmail)
		return
	}
	//Hash password
	hashedPassword, err := utils.HashPassword(payload.Password)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("could not hash password: %w", err))
		return
	}
	//Check ref_code
	var referrerID *int
	if payload.RefCode != "" {
		referrer, err := h.store.GetUserByReferralCode(payload.RefCode)
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid referral code: %w", err))
			return
		}
		referrerID = &referrer.ID
	}
	log.Printf("Referrer ID: %v", referrerID)

	//Save User
	newUser := types.User{
		Email:      payload.Email,
		Password:   hashedPassword,
		ReferrerID: referrerID,
	}
	log.Printf("Saving user: %+v", newUser)

	if err := h.store.SaveUser(newUser); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("could not create user: %w", err))
		return
	}
	utils.WriteJSON(w, http.StatusCreated, map[string]string{"message": "user registered successfully"})

}
