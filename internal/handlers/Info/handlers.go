package info

import (
	"encoding/json"
	"fmt"
	"net/http"

	"RefCode.com/m/types"
	"RefCode.com/m/utils"
)

type Handler struct {
	store types.UserStore
}

func NewHandler(store types.UserStore) *Handler {
	return &Handler{store: store}
}

func (h *Handler) GetRefCodeByEmail(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Email == "" {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("valid email is required"))
		return
	}

	user, err := h.store.GetUserByEmail(req.Email)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("could not get user: %w", err))
		return
	}

	if user.ReferralCode == nil || *user.ReferralCode == "" {
		utils.WriteError(w, http.StatusNotFound, fmt.Errorf("user does not have a referral code"))
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{"referral_code": *user.ReferralCode})
}

func (h *Handler) GetReferralsByIdReferrer(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ReferrerID int `json:"referrer_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ReferrerID == 0 {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("valid referrer_id is required"))
		return
	}

	referrals, err := h.store.GetReferralsByReferrerID(req.ReferrerID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("could not get referrals: %w", err))
		return
	}

	if len(referrals) == 0 {
		utils.WriteJSON(w, http.StatusOK, map[string][]string{"emails": {}})
		return
	}

	var emails []string
	for _, ref := range referrals {
		emails = append(emails, ref.Email)
	}

	utils.WriteJSON(w, http.StatusOK, map[string][]string{"emails": emails})
}
