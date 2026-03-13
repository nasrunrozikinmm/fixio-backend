package seeder

import (
	"log"

	models "fixio/internal/modules/auth/entity"
	sectorModels "fixio/internal/modules/sector/entity"
	"fixio/pkg/config"
	"fixio/pkg/helpers"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SeedDefaultAdmin creates a default admin user if none exists
func SeedDefaultAdmin(db *gorm.DB, env *config.Environment) {
	var count int64
	db.Model(&models.User{}).Where("role = ?", helpers.RoleAdministrator).Count(&count)
	if count > 0 {
		return
	}

	admin := &models.User{
		ID:         uuid.New(),
		Name:       "Administrator",
		Email:      "mapokace@gmail.com",
		Role:       helpers.RoleAdministrator,
		Provider:   helpers.ProviderGoogle,
		ProviderID: "google-admin-001",
	}

	if err := db.Create(admin).Error; err != nil {
		log.Printf("⚠️  Failed to seed admin: %v", err)
		return
	}
	log.Println("✅ Default admin user seeded")
}

// SeedDefaultSectors creates default policy sectors if none exist
func SeedDefaultSectors(db *gorm.DB) {
	var count int64
	db.Model(&sectorModels.Sector{}).Count(&count)
	if count > 0 {
		return
	}

	sectors := []sectorModels.Sector{
		{ID: uuid.New(), Name: "Pendidikan", Slug: "pendidikan"},
		{ID: uuid.New(), Name: "Kesehatan", Slug: "kesehatan"},
		{ID: uuid.New(), Name: "Infrastruktur", Slug: "infrastruktur"},
		{ID: uuid.New(), Name: "Ekonomi", Slug: "ekonomi"},
		{ID: uuid.New(), Name: "Hukum", Slug: "hukum"},
		{ID: uuid.New(), Name: "Lingkungan", Slug: "lingkungan"},
		{ID: uuid.New(), Name: "Sosial", Slug: "sosial"},
		{ID: uuid.New(), Name: "Pertahanan & Keamanan", Slug: "pertahanan-keamanan"},
		{ID: uuid.New(), Name: "Teknologi", Slug: "teknologi"},
		{ID: uuid.New(), Name: "Transportasi", Slug: "transportasi"},
		{ID: uuid.New(), Name: "Energi", Slug: "energi"},
		{ID: uuid.New(), Name: "Pertanian", Slug: "pertanian"},
	}

	for _, s := range sectors {
		if err := db.Create(&s).Error; err != nil {
			log.Printf("⚠️  Failed to seed sector '%s': %v", s.Name, err)
		}
	}
	log.Printf("✅ %d default sectors seeded", len(sectors))
}
