package auth

import (
	"github.com/skuid/spec/mapvalue"
)

// PlinyUser stores information from a pliny auth provider
type PlinyUser struct {
	ID                     string              `json:"id"`
	FirstName              string              `json:"firstName"`
	LastName               string              `json:"lastName"`
	Email                  string              `json:"email"`
	Username               string              `json:"username"`
	FederationID           string              `json:"federationId"`
	SiteID                 string              `json:"siteId"`
	ProfileName            string              `json:"profileName"`
	Subdomain              string              `json:"subdomain"`
	NamedPermissions       map[string][]string `json:"namedPermissions"`
	IdentityProviderClaims map[string][]string `json:"identityProviderClaims"`
	SessionVariables       map[string]string   `json:"sessionVariables"`
	FeatureFlags           map[string]bool     `json:"featureFlags"`
}

// UserInfo is used when data source conditions need to evaluate current user values.
type UserInfo interface {
	IsAdmin() bool
	IsSkuidAdmin() bool
	GetFieldValue(string) (string, bool)
	GetIdentityProviderClaim(string) ([]string, bool)
	GetIdentityAttributeName(string) string
	GetProfileName() string
	GetFeatureFlags() map[string]bool
}

// IsAdmin returns whether or not this user has admin privileges
func (p PlinyUser) IsAdmin() bool {
	skuidNamedPermissions, hasSkuidNamedPermissions := p.NamedPermissions["skuid"]
	if hasSkuidNamedPermissions {
		return mapvalue.StringSliceContainsKey(skuidNamedPermissions, "configure_site")
	}
	// Once the version of pliny that returns named permissions in the auth check
	// makes it to production, this can be removed and we can just return false
	return p.ProfileName == "Admin"
}

// IsSkuidAdmin returns whether or not this user is a Skuid admin (i.e. DMV)
func (p PlinyUser) IsSkuidAdmin() bool {
	skuidNamedPermissions, hasSkuidNamedPermissions := p.NamedPermissions["skuid"]
	if hasSkuidNamedPermissions {
		return mapvalue.StringSliceContainsKey(skuidNamedPermissions, "skuid_admin")
	}
	return false
}

// GetFieldValue retrieves a particular value from userinfo by field name
func (p PlinyUser) GetFieldValue(field string) (string, bool) {
	switch field {
	case "first_name":
		return p.FirstName, true
	case "last_name":
		return p.LastName, true
	case "email":
		return p.Email, true
	case "username":
		return p.Username, true
	case "user_id":
		return p.ID, true
	case "federation_id":
		return p.FederationID, true
	case "site_id":
		return p.SiteID, true
	case "profile_name":
		return p.ProfileName, true
	case "subdomain":
		return p.Subdomain, true
	default:
		return "", false
	}
}

// GetIdentityProviderClaim retrieves a particular claim
// returned from the user's session's identity provider,
// if the user's session was created using SAML.
func (p PlinyUser) GetIdentityProviderClaim(claimName string) ([]string, bool) {

	idpClaims := p.IdentityProviderClaims

	if idpClaims != nil {
		idpClaim := idpClaims[claimName]

		if len(idpClaim) > 0 {
			return idpClaim, true
		}
	}

	return []string{}, false

}

// GetIdentityAttributeName retrieves the saml attribute name tied to the session variable name
func (p PlinyUser) GetIdentityAttributeName(sessionVariableName string) string {
	sessionVariables := p.SessionVariables
	claimName := ""

	if sessionVariables != nil {
		claimName = sessionVariables[sessionVariableName]
	}
	return claimName
}

func (p PlinyUser) GetProfileName() string {
	return p.ProfileName
}

func (p PlinyUser) GetFeatureFlags() map[string]bool {
	return p.FeatureFlags
}
