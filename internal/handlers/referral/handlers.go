package referral

import (
	"fmt"
	"net/http"
	"time"

	"RefCode.com/m/types"
	"RefCode.com/m/utils"
	"RefCode.com/m/utils/generate"
)

type Handler struct {
	store types.UserStore
}

func NewHandler(store types.UserStore) *Handler {
	return &Handler{store: store}
}

func (h *Handler) CreateRefCode(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(int)

	user, err := h.store.GetUserById(userID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("could not get user: %w", err))
		return
	}

	if user.ReferralCode != nil {
		utils.WriteError(w, http.StatusConflict, fmt.Errorf("user already has a referral code"))
		return
	}

	refCode := generate.GenerateReferralCode()
	expiry := time.Now().Add(24 * time.Hour) //

	if err := h.store.SaveReferralCode(userID, refCode, expiry); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("could not save referral code: %w", err))
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{"referral_code": refCode})
}

func (h *Handler) DeleteRefCode(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(int)

	user, err := h.store.GetUserById(userID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("could not get user: %w", err))
		return
	}

	if user.ReferralCode == nil || *user.ReferralCode == "" {
		utils.WriteError(w, http.StatusNotFound, fmt.Errorf("user does not have a referral code"))
		return
	}

	if err := h.store.SaveReferralCode(userID, "", time.Time{}); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("could not delete referral code: %w", err))
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{"message": "referral code deleted successfully"})
}
