package controllers

import (
	"encoding/json"
	"fmt"
	"io"

	"fixio/internal/modules/auth/dto"
	services "fixio/internal/modules/auth/usecase"
	"fixio/pkg/config"
	"fixio/pkg/network"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
	googleOAuth "golang.org/x/oauth2/google"
)

type authController struct {
	network.BaseController
	service      services.AuthService
	googleConfig *oauth2.Config
	env          *config.Environment
}

// NewAuthController creates a new auth controller
func NewAuthController(
	authFn network.AuthenticationProvider,
	authzFn network.AuthorizationProvider,
	service services.AuthService,
	env *config.Environment,
) network.Controller {
	googleConfig := &oauth2.Config{
		ClientID:     env.GoogleClientID,
		ClientSecret: env.GoogleClientSecret,
		RedirectURL:  env.GoogleRedirectURL,
		Scopes:       []string{"openid", "profile", "email"},
		Endpoint:     googleOAuth.Endpoint,
	}

	return &authController{
		BaseController: network.NewBaseController("/auth", authFn, authzFn),
		service:        service,
		googleConfig:   googleConfig,
		env:            env,
	}
}

// MountRoutes registers all auth routes
func (c *authController) MountRoutes(rg *fiber.Group) {
	rg.Get("/google", c.GoogleLogin)
	rg.Get("/google/callback", c.GoogleCallback)
	rg.Get("/me", c.Authentication(), c.GetMe)
	rg.Put("/me", c.Authentication(), c.UpdateProfile)
	rg.Post("/logout", c.Logout)
}

// GoogleLogin godoc
// @Summary     Login via Google
// @Description Redirect ke Google OAuth consent screen untuk autentikasi
// @Tags        Auth
// @Produce     json
// @Success     302 {string} string "Redirect ke Google"
// @Router      /api/auth/google [get]
func (c *authController) GoogleLogin(ctx *fiber.Ctx) error {
	url := c.googleConfig.AuthCodeURL("state", oauth2.AccessTypeOffline)
	return ctx.Redirect(url)
}

// GoogleCallback godoc
// @Summary     Google OAuth Callback
// @Description Handle callback dari Google OAuth, buat/login user, set cookie, dan redirect ke frontend
// @Tags        Auth
// @Accept      json
// @Produce     json
// @Param       code query string true "Authorization code dari Google"
// @Success     302 {string} string "Redirect ke frontend dengan token"
// @Failure     400 {object} network.ErrorResponse "Kode otorisasi tidak ditemukan"
// @Failure     500 {object} network.ErrorResponse "Gagal mendapatkan info user"
// @Router      /api/auth/google/callback [get]
func (c *authController) GoogleCallback(ctx *fiber.Ctx) error {
	code := ctx.Query("code")
	if code == "" {
		return c.Send(ctx).BadRequestError("Kode otorisasi tidak ditemukan", nil)
	}

	token, err := c.googleConfig.Exchange(ctx.Context(), code)
	if err != nil {
		return c.Send(ctx).BadRequestError("Gagal menukar kode otorisasi", err)
	}

	// Get user info from Google
	client := c.googleConfig.Client(ctx.Context(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return c.Send(ctx).InternalServerError("Gagal mendapatkan info user dari Google", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var googleUser struct {
		ID      string `json:"id"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
	}
	json.Unmarshal(body, &googleUser)

	userInfo := &dto.OAuthUserInfo{
		ID:        googleUser.ID,
		Email:     googleUser.Email,
		Name:      googleUser.Name,
		AvatarURL: googleUser.Picture,
		Provider:  "google",
	}

	authResp, err := c.service.HandleOAuthCallback(ctx.Context(), userInfo)
	if err != nil {
		return c.Send(ctx).HandleError(err)
	}

	// Set HTTP-only cookie
	ctx.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    authResp.Token,
		HTTPOnly: true,
		Secure:   c.env.IsProduction(),
		SameSite: "Lax",
		Path:     "/",
	})

	// Redirect to frontend
	redirectURL := fmt.Sprintf("%s/auth/callback?token=%s", c.env.FrontendURL, authResp.Token)
	return ctx.Redirect(redirectURL)
}

// GetMe godoc
// @Summary     Get profil user saat ini
// @Description Mengambil data profil user yang sedang login berdasarkan JWT token
// @Tags        Auth
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Success     200 {object} network.Response{data=dto.UserProfile} "Profil user berhasil diambil"
// @Failure     401 {object} network.ErrorResponse "Unauthorized"
// @Router      /api/auth/me [get]
func (c *authController) GetMe(ctx *fiber.Ctx) error {
	userIDStr, _ := ctx.Locals("userId").(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.Send(ctx).UnauthorizedError("User ID tidak valid", err)
	}

	profile, err := c.service.GetCurrentUser(ctx.Context(), userID)
	if err != nil {
		return c.Send(ctx).HandleError(err)
	}

	return c.Send(ctx).SuccessDataResponse("Profil user berhasil diambil", profile)
}

// UpdateProfile godoc
// @Summary     Update profil user
// @Description Mengupdate profil user yang sedang login (nama, bio, lokasi)
// @Tags        Auth
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       body body dto.UpdateProfileRequest true "Data profil yang akan diupdate"
// @Success     200 {object} network.Response{data=dto.UserProfile} "Profil berhasil diupdate"
// @Failure     400 {object} network.ErrorResponse "Request body tidak valid"
// @Failure     401 {object} network.ErrorResponse "Unauthorized"
// @Router      /api/auth/me [put]
func (c *authController) UpdateProfile(ctx *fiber.Ctx) error {
	userIDStr, _ := ctx.Locals("userId").(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.Send(ctx).UnauthorizedError("User ID tidak valid", err)
	}

	req := new(dto.UpdateProfileRequest)
	if err := ctx.BodyParser(req); err != nil {
		return c.Send(ctx).BadRequestError(network.ErrInvalidBody, err)
	}
	if result := network.Validate(req); !result.Valid {
		return c.Send(ctx).BadRequestError(result.FirstMessage(), nil)
	}

	profile, err := c.service.UpdateProfile(ctx.Context(), userID, req)
	if err != nil {
		return c.Send(ctx).HandleError(err)
	}

	return c.Send(ctx).SuccessDataResponse("Profil berhasil diupdate", profile)
}

// Logout godoc
// @Summary     Logout user
// @Description Menghapus cookie access_token untuk logout user
// @Tags        Auth
// @Produce     json
// @Success     200 {object} network.Response "Logout berhasil"
// @Router      /api/auth/logout [post]
func (c *authController) Logout(ctx *fiber.Ctx) error {
	ctx.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    "",
		HTTPOnly: true,
		Secure:   c.env.IsProduction(),
		SameSite: "Lax",
		Path:     "/",
		MaxAge:   -1,
	})
	return c.Send(ctx).SuccessMsgResponse("Logout berhasil")
}
