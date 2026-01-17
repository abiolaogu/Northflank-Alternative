// Package main is a simplified local development server
// that uses SQLite for easy local testing without external dependencies.
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

// Project model
type Project struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Services    int       `json:"services"`
	Databases   int       `json:"databases"`
	Status      string    `json:"status"`
	Members     int       `json:"members"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// In-memory store with file persistence
type Store struct {
	mu       sync.RWMutex
	projects map[string]*Project
	filePath string
}

func NewStore(filePath string) *Store {
	s := &Store{
		projects: make(map[string]*Project),
		filePath: filePath,
	}
	s.load()
	return s
}

func (s *Store) load() {
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		// Initialize with default projects
		s.projects["production-api"] = &Project{
			ID: "1", Name: "production-api", Description: "Main production API",
			Services: 8, Databases: 2, Status: "healthy", Members: 5,
			CreatedAt: time.Now().Add(-14 * 24 * time.Hour), UpdatedAt: time.Now(),
		}
		s.projects["staging-env"] = &Project{
			ID: "2", Name: "staging-env", Description: "Staging environment",
			Services: 4, Databases: 1, Status: "healthy", Members: 3,
			CreatedAt: time.Now().Add(-30 * 24 * time.Hour), UpdatedAt: time.Now(),
		}
		s.save()
		return
	}
	var projects []*Project
	if err := json.Unmarshal(data, &projects); err == nil {
		for _, p := range projects {
			s.projects[p.Name] = p
		}
	}
}

func (s *Store) save() {
	s.mu.RLock()
	defer s.mu.RUnlock()

	projects := make([]*Project, 0, len(s.projects))
	for _, p := range s.projects {
		projects = append(projects, p)
	}

	data, _ := json.MarshalIndent(projects, "", "  ")
	os.WriteFile(s.filePath, data, 0644)
}

func (s *Store) List() []*Project {
	s.mu.RLock()
	defer s.mu.RUnlock()

	projects := make([]*Project, 0, len(s.projects))
	for _, p := range s.projects {
		projects = append(projects, p)
	}
	return projects
}

func (s *Store) Get(name string) *Project {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.projects[name]
}

func (s *Store) Create(p *Project) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.projects[p.Name]; exists {
		return fmt.Errorf("project %s already exists", p.Name)
	}

	p.ID = fmt.Sprintf("%d", time.Now().UnixNano())
	p.CreatedAt = time.Now()
	p.UpdatedAt = time.Now()
	p.Status = "healthy"
	p.Services = 0
	p.Databases = 0
	p.Members = 1

	s.projects[p.Name] = p
	go s.save()
	return nil
}

func (s *Store) Delete(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.projects[name]; !exists {
		return fmt.Errorf("project %s not found", name)
	}

	delete(s.projects, name)
	go s.save()
	return nil
}

// CORS middleware
func cors(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next(w, r)
	}
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	store := NewStore("skyforge_data.json")

	// Health check
	http.HandleFunc("/health", cors(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))

	// List projects
	http.HandleFunc("/api/v1/projects", cors(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.Method {
		case "GET":
			projects := store.List()
			json.NewEncoder(w).Encode(projects)

		case "POST":
			var p Project
			if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if err := store.Create(&p); err != nil {
				http.Error(w, err.Error(), http.StatusConflict)
				return
			}
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(p)

		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	// Single project operations
	http.HandleFunc("/api/v1/projects/", cors(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		name := r.URL.Path[len("/api/v1/projects/"):]

		switch r.Method {
		case "GET":
			p := store.Get(name)
			if p == nil {
				http.Error(w, "Not found", http.StatusNotFound)
				return
			}
			json.NewEncoder(w).Encode(p)

		case "DELETE":
			if err := store.Delete(name); err != nil {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}
			w.WriteHeader(http.StatusNoContent)

		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	log.Printf("🚀 SkyForge API running on http://localhost:%s", port)
	log.Printf("📁 Data persisted to skyforge_data.json")
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
