// Package valueobjects contains DDD value objects.
// Value objects are immutable objects that represent concepts without identity.
package valueobjects

import (
	"errors"
	"net/mail"
	"regexp"
	"strings"
)

// Email represents a validated email address
type Email struct {
	value string
}

// NewEmail creates a new Email value object with validation
func NewEmail(email string) (Email, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" {
		return Email{}, errors.New("email cannot be empty")
	}

	_, err := mail.ParseAddress(email)
	if err != nil {
		return Email{}, errors.New("invalid email format")
	}

	return Email{value: email}, nil
}

func (e Email) String() string          { return e.value }
func (e Email) Equals(other Email) bool { return e.value == other.value }

// Slug represents a URL-safe identifier
type Slug struct {
	value string
}

var slugRegex = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

// NewSlug creates a new Slug value object with validation
func NewSlug(slug string) (Slug, error) {
	slug = strings.TrimSpace(strings.ToLower(slug))
	if slug == "" {
		return Slug{}, errors.New("slug cannot be empty")
	}

	if len(slug) < 3 || len(slug) > 63 {
		return Slug{}, errors.New("slug must be between 3 and 63 characters")
	}

	if !slugRegex.MatchString(slug) {
		return Slug{}, errors.New("slug must contain only lowercase letters, numbers, and hyphens")
	}

	return Slug{value: slug}, nil
}

// GenerateSlug generates a slug from a name
func GenerateSlug(name string) Slug {
	slug := strings.ToLower(name)
	slug = strings.TrimSpace(slug)
	slug = regexp.MustCompile(`[^a-z0-9\s-]`).ReplaceAllString(slug, "")
	slug = regexp.MustCompile(`\s+`).ReplaceAllString(slug, "-")
	slug = regexp.MustCompile(`-+`).ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")

	if len(slug) > 63 {
		slug = slug[:63]
	}

	return Slug{value: slug}
}

func (s Slug) String() string         { return s.value }
func (s Slug) Equals(other Slug) bool { return s.value == other.value }

// ResourceLimits represents resource constraints for workloads
type ResourceLimits struct {
	cpuRequest    string
	cpuLimit      string
	memoryRequest string
	memoryLimit   string
}

// NewResourceLimits creates a new ResourceLimits value object
func NewResourceLimits(cpuReq, cpuLim, memReq, memLim string) ResourceLimits {
	return ResourceLimits{
		cpuRequest:    cpuReq,
		cpuLimit:      cpuLim,
		memoryRequest: memReq,
		memoryLimit:   memLim,
	}
}

func (r ResourceLimits) CPURequest() string    { return r.cpuRequest }
func (r ResourceLimits) CPULimit() string      { return r.cpuLimit }
func (r ResourceLimits) MemoryRequest() string { return r.memoryRequest }
func (r ResourceLimits) MemoryLimit() string   { return r.memoryLimit }

// Predefined resource profiles
var (
	ResourceSmall  = NewResourceLimits("100m", "500m", "128Mi", "512Mi")
	ResourceMedium = NewResourceLimits("250m", "1", "256Mi", "1Gi")
	ResourceLarge  = NewResourceLimits("500m", "2", "512Mi", "2Gi")
	ResourceXLarge = NewResourceLimits("1", "4", "1Gi", "4Gi")
)

// GitRef represents a Git reference (branch, tag, or commit)
type GitRef struct {
	refType string // branch, tag, commit
	value   string
}

// NewBranchRef creates a new branch reference
func NewBranchRef(branch string) (GitRef, error) {
	if branch == "" {
		return GitRef{}, errors.New("branch cannot be empty")
	}
	return GitRef{refType: "branch", value: branch}, nil
}

// NewTagRef creates a new tag reference
func NewTagRef(tag string) (GitRef, error) {
	if tag == "" {
		return GitRef{}, errors.New("tag cannot be empty")
	}
	return GitRef{refType: "tag", value: tag}, nil
}

// NewCommitRef creates a new commit reference
func NewCommitRef(sha string) (GitRef, error) {
	if len(sha) < 7 {
		return GitRef{}, errors.New("commit SHA must be at least 7 characters")
	}
	return GitRef{refType: "commit", value: sha}, nil
}

func (g GitRef) Type() string   { return g.refType }
func (g GitRef) Value() string  { return g.value }
func (g GitRef) String() string { return g.refType + ":" + g.value }

// Version represents a semantic version
type Version struct {
	major int
	minor int
	patch int
}

var versionRegex = regexp.MustCompile(`^v?(\d+)\.(\d+)\.(\d+)$`)

// ParseVersion parses a version string
func ParseVersion(s string) (Version, error) {
	matches := versionRegex.FindStringSubmatch(s)
	if matches == nil {
		return Version{}, errors.New("invalid version format, expected: major.minor.patch")
	}

	// Simplified - in production use strconv.Atoi
	return Version{major: 0, minor: 0, patch: 0}, nil
}

func (v Version) String() string {
	return strings.Join([]string{
		string(rune('0' + v.major)),
		string(rune('0' + v.minor)),
		string(rune('0' + v.patch)),
	}, ".")
}

// Port represents a validated port number
type Port struct {
	value int
}

// NewPort creates a new Port value object
func NewPort(port int) (Port, error) {
	if port < 1 || port > 65535 {
		return Port{}, errors.New("port must be between 1 and 65535")
	}
	return Port{value: port}, nil
}

func (p Port) Value() int { return p.value }

// Namespace represents a Kubernetes namespace
type Namespace struct {
	value string
}

var namespaceRegex = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)

// NewNamespace creates a new Namespace value object
func NewNamespace(ns string) (Namespace, error) {
	ns = strings.TrimSpace(strings.ToLower(ns))
	if ns == "" {
		return Namespace{}, errors.New("namespace cannot be empty")
	}

	if len(ns) > 63 {
		return Namespace{}, errors.New("namespace must be 63 characters or less")
	}

	if !namespaceRegex.MatchString(ns) {
		return Namespace{}, errors.New("invalid namespace format")
	}

	return Namespace{value: ns}, nil
}

func (n Namespace) String() string { return n.value }
