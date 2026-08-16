package httpapi

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"math/big"
	"net/http"
	"strings"
	"time"

	"loombloom/pkg/store"
	"loombloom/pkg/views"
)

type contextKey string

const orgContextKey contextKey = "organization"

// SessionCookieName is the name of the session cookie
const SessionCookieName = "session_token"

// GetOrg extracts the organization from the context.
func GetOrg(ctx context.Context) (store.Organization, bool) {
	org, ok := ctx.Value(orgContextKey).(store.Organization)
	return org, ok
}

// GenerateSessionToken creates a secure random token
func GenerateSessionToken() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// GenerateOTP produces a 6-digit random code
func GenerateOTP() string {
	n, _ := rand.Int(rand.Reader, big.NewInt(900000))
	return n.Add(n, big.NewInt(100000)).String()
}

// OnboardingMiddleware checks the user session and handles redirects based on the onboarding state
func (api API) OnboardingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// Bypasses for static and health endpoints
		if path == "/healthz" || path == "/readyz" || path == "/setup" || path == "/login" {
			next.ServeHTTP(w, r)
			return
		}

		// Read session token
		cookie, err := r.Cookie(SessionCookieName)
		if err != nil {
			// No session. If trying to access onboarding pages, let through
			if path == "/verify" || path == "/plans" || path == "/checkout" || path == "/payment/checkout" || path == "/payment/callback" {
				next.ServeHTTP(w, r)
				return
			}
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Fetch session from database
		sess, err := api.store.GetSession(r.Context(), cookie.Value)
		if err != nil {
			// Invalid or expired session
			api.clearSessionCookie(w)
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Fetch organization
		org, err := api.store.GetOrganization(r.Context(), sess.OrganizationID)
		if err != nil {
			api.clearSessionCookie(w)
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Check onboarding state
		if !org.PhoneVerified {
			if path != "/verify" && path != "/verify/resend" {
				http.Redirect(w, r, "/verify?org_id="+org.ID, http.StatusSeeOther)
				return
			}
		} else if org.SubscriptionPlan == "" || org.SubscriptionPlan == "inactive" {
			if path != "/plans" && path != "/checkout" && path != "/payment/checkout" && path != "/payment/callback" && path != "/verify" {
				http.Redirect(w, r, "/plans", http.StatusSeeOther)
				return
			}
		} else {
			// Fully onboarded. If trying to go back to onboarding/auth pages, redirect to dashboard
			if path == "/setup" || path == "/login" || path == "/verify" || path == "/plans" {
				http.Redirect(w, r, "/", http.StatusSeeOther)
				return
			}
		}

		// Inject organization into request context
		ctx := context.WithValue(r.Context(), orgContextKey, org)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (api API) clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   false, // Set to true in production with TLS
		SameSite: http.SameSiteLaxMode,
	})
}

// normalizePhone removes spaces/dashes and prepends +91 to 10-digit numbers
func normalizePhone(phone string) string {
	phone = strings.ReplaceAll(phone, " ", "")
	phone = strings.ReplaceAll(phone, "-", "")
	if len(phone) == 10 && !strings.HasPrefix(phone, "+") {
		return "+91" + phone
	}
	return phone
}

// --- Handlers ---

// setupPageHandler handles GET and POST for organization registration
func (api API) setupPageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = views.SetupPage("").Render(r.Context(), w)
		return
	}

	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_ = views.SetupPage("Failed to parse form: "+err.Error()).Render(r.Context(), w)
			return
		}

		name := r.FormValue("name")
		ownerName := r.FormValue("owner_name")
		email := r.FormValue("owner_email")
		phone := normalizePhone(r.FormValue("phone"))

		if name == "" || ownerName == "" || email == "" || phone == "" {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_ = views.SetupPage("All fields are required.").Render(r.Context(), w)
			return
		}

		api.logger.Info("New registration attempt", "enterprise", name, "email", email, "phone", phone)

		// Create organization with unverified status
		org, err := api.store.CreateOrganization(r.Context(), store.Organization{
			Name:             name,
			OwnerName:        ownerName,
			OwnerEmail:       email,
			Phone:            phone,
			PhoneVerified:    false,
			SubscriptionPlan: "inactive",
		})
		if err != nil {
			api.logger.Error("failed to create organization", "error", err)
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_ = views.SetupPage("Failed to create enterprise: "+err.Error()).Render(r.Context(), w)
			return
		}

		api.logger.Info("Organization registered successfully", "org_id", org.ID, "enterprise", name)

		// Generate OTP and save
		otp := GenerateOTP()
		expiry := time.Now().Add(10 * time.Minute)
		if err := api.store.UpdateOTP(r.Context(), org.ID, otp, expiry); err != nil {
			api.logger.Error("failed to save OTP", "error", err)
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_ = views.SetupPage("Failed to generate OTP: "+err.Error()).Render(r.Context(), w)
			return
		}

		// LOG OTP FOR LOCAL DEVELOPMENT (Mock SMS provider)
		api.logger.Info("-------------------------------------------")
		api.logger.Info("SMS DISPATCHED", "phone", phone, "otp_code", otp)
		api.logger.Info("-------------------------------------------")

		http.Redirect(w, r, "/verify?org_id="+org.ID, http.StatusSeeOther)
	}
}

// loginPageHandler handles GET and POST for OTP-based login
func (api API) loginPageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = views.LoginPage("").Render(r.Context(), w)
		return
	}

	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_ = views.LoginPage("Failed to parse form: "+err.Error()).Render(r.Context(), w)
			return
		}

		email := strings.TrimSpace(r.FormValue("email"))
		phone := normalizePhone(strings.TrimSpace(r.FormValue("phone")))
		
		if email == "" && phone == "" {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_ = views.LoginPage("Please enter your email or phone number.").Render(r.Context(), w)
			return
		}

		api.logger.Info("Login attempt", "email", email, "phone", phone)

		var org store.Organization
		var err error
		if email != "" {
			org, err = api.store.GetOrganizationByEmail(r.Context(), email)
		} else {
			org, err = api.store.GetOrganizationByPhone(r.Context(), phone)
		}

		if err != nil {
			api.logger.Warn("Login failed: no account found", "email", email, "phone", phone)
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_ = views.LoginPage("No account found with that email/phone. Please register first.").Render(r.Context(), w)
			return
		}

		api.logger.Info("Login: account found", "org_id", org.ID, "enterprise", org.Name)

		// Generate OTP and save
		otp := GenerateOTP()
		expiry := time.Now().Add(10 * time.Minute)
		if err := api.store.UpdateOTP(r.Context(), org.ID, otp, expiry); err != nil {
			api.logger.Error("failed to save login OTP", "error", err)
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_ = views.LoginPage("Failed to send OTP. Please try again.").Render(r.Context(), w)
			return
		}

		// LOG OTP FOR LOCAL DEVELOPMENT (Mock SMS provider)
		api.logger.Info("-------------------------------------------")
		api.logger.Info("LOGIN OTP DISPATCHED", "phone", org.Phone, "otp_code", otp)
		api.logger.Info("-------------------------------------------")

		http.Redirect(w, r, "/verify?org_id="+org.ID, http.StatusSeeOther)
	}
}

// verifyPageHandler handles phone verification via OTP
func (api API) verifyPageHandler(w http.ResponseWriter, r *http.Request) {
	orgID := r.URL.Query().Get("org_id")
	if orgID == "" {
		// Try to look it up from session context
		if org, ok := GetOrg(r.Context()); ok {
			orgID = org.ID
		}
	}

	if r.Method == http.MethodGet {
		org, err := api.store.GetOrganization(r.Context(), orgID)
		if err != nil {
			http.Redirect(w, r, "/setup", http.StatusSeeOther)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = views.VerifyPage(org.ID, org.Phone, "").Render(r.Context(), w)
		return
	}

	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_ = views.VerifyPage(orgID, "", "Failed to parse code.").Render(r.Context(), w)
			return
		}

		orgID = r.FormValue("org_id")
		org, err := api.store.GetOrganization(r.Context(), orgID)
		if err != nil {
			http.Redirect(w, r, "/setup", http.StatusSeeOther)
			return
		}

		// Rebuild 6-digit OTP
		otp := r.FormValue("otp_1") + r.FormValue("otp_2") + r.FormValue("otp_3") + r.FormValue("otp_4") + r.FormValue("otp_5") + r.FormValue("otp_6")

		ok, err := api.store.VerifyOTP(r.Context(), orgID, otp)
		if err != nil {
			api.logger.Error("verify OTP error", "error", err)
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_ = views.VerifyPage(orgID, org.Phone, "Verification system error").Render(r.Context(), w)
			return
		}

		if !ok {
			api.logger.Warn("OTP verification failed: invalid or expired code", "org_id", orgID)
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_ = views.VerifyPage(orgID, org.Phone, "Invalid or expired OTP code.").Render(r.Context(), w)
			return
		}

		api.logger.Info("OTP verified successfully", "org_id", orgID)

		// Extend session cookie now that they are verified
		token := GenerateSessionToken()
		_, err = api.store.CreateSession(r.Context(), orgID, token, time.Now().Add(24*time.Hour))
		if err != nil {
			api.logger.Error("create verified session error", "error", err)
		}

		http.SetCookie(w, &http.Cookie{
			Name:     SessionCookieName,
			Value:    token,
			Path:     "/",
			Expires:  time.Now().Add(24 * time.Hour),
			HttpOnly: true,
			Secure:   false,
			SameSite: http.SameSiteLaxMode,
		})

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

// verifyResendHandler handles resending SMS OTP
func (api API) verifyResendHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	_ = r.ParseForm()
	orgID := r.FormValue("org_id")
	org, err := api.store.GetOrganization(r.Context(), orgID)
	if err != nil {
		http.Redirect(w, r, "/setup", http.StatusSeeOther)
		return
	}

	otp := GenerateOTP()
	expiry := time.Now().Add(10 * time.Minute)
	if err := api.store.UpdateOTP(r.Context(), org.ID, otp, expiry); err != nil {
		api.logger.Error("failed to generate resend OTP", "error", err)
		http.Redirect(w, r, "/verify?org_id="+orgID, http.StatusSeeOther)
		return
	}

	// LOG OTP FOR LOCAL DEVELOPMENT
	api.logger.Info("-------------------------------------------")
	api.logger.Info("SMS RESENT", "phone", org.Phone, "otp_code", otp)
	api.logger.Info("-------------------------------------------")

	http.Redirect(w, r, "/verify?org_id="+orgID, http.StatusSeeOther)
}

// plansPageHandler renders the pricing grid
func (api API) plansPageHandler(w http.ResponseWriter, r *http.Request) {
	org, ok := GetOrg(r.Context())
	if !ok {
		http.Redirect(w, r, "/setup", http.StatusSeeOther)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = views.PlansPage(org, "").Render(r.Context(), w)
}

// checkoutHandler initiates checkout by redirecting to payment screen
func (api API) checkoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/plans", http.StatusSeeOther)
		return
	}

	_ = r.ParseForm()
	plan := r.FormValue("plan")
	if plan == "" {
		plan = "starter"
	}

	// Redirect to mock payment gateway screen
	http.Redirect(w, r, "/payment/checkout?plan="+plan, http.StatusSeeOther)
}

// mockCheckoutPageHandler renders the mock Stripe/Razorpay payment interface
func (api API) mockCheckoutPageHandler(w http.ResponseWriter, r *http.Request) {
	plan := r.URL.Query().Get("plan")
	org, ok := GetOrg(r.Context())
	if !ok {
		http.Redirect(w, r, "/setup", http.StatusSeeOther)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = views.MockCheckoutPage(org, plan).Render(r.Context(), w)
}

// mockPaymentCallbackHandler processes successful payment response
func (api API) mockPaymentCallbackHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/plans", http.StatusSeeOther)
		return
	}

	_ = r.ParseForm()
	plan := r.FormValue("plan")
	orgID := r.FormValue("org_id")

	if plan == "" || orgID == "" {
		http.Redirect(w, r, "/plans", http.StatusSeeOther)
		return
	}

	// Update organization plan status in DB
	err := api.store.UpdateSubscriptionPlan(r.Context(), orgID, plan)
	if err != nil {
		api.logger.Error("failed to activate subscription plan", "error", err)
		http.Redirect(w, r, "/plans", http.StatusSeeOther)
		return
	}

	api.logger.Info("Payment successful. Subscription activated.", "org_id", orgID, "plan", plan)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// logoutHandler clears session cookies
func (api API) logoutHandler(w http.ResponseWriter, r *http.Request) {
	org, _ := GetOrg(r.Context())
	cookie, err := r.Cookie(SessionCookieName)
	if err == nil {
		_ = api.store.DeleteSession(r.Context(), cookie.Value)
	}
	api.logger.Info("User logged out", "org_id", org.ID, "enterprise", org.Name)
	api.clearSessionCookie(w)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
