package seeder

import (
	"log"
	"time"

	userModels "fixio/internal/modules/auth/entity"
	commentModels "fixio/internal/modules/comment/entity"
	postModels "fixio/internal/modules/post/entity"
	regionModels "fixio/internal/modules/region/entity"
	sectorModels "fixio/internal/modules/sector/entity"
	voteModels "fixio/internal/modules/vote/entity"
	"fixio/pkg/helpers"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SeedDefaultRegions creates default regions if none exist (Nasional → Provinsi → Kota)
func SeedDefaultRegions(db *gorm.DB) {
	var count int64
	db.Model(&regionModels.Region{}).Count(&count)
	if count > 0 {
		return
	}

	nasionalID := uuid.New()
	dkiID := uuid.New()
	jabarID := uuid.New()
	jatimID := uuid.New()

	regions := []regionModels.Region{
		// Level 1: Nasional
		{ID: nasionalID, Name: "Nasional", Slug: "nasional", Type: "nasional", ParentID: nil},
		// Level 2: Provinsi
		{ID: dkiID, Name: "DKI Jakarta", Slug: "dki-jakarta", Type: "provinsi", ParentID: &nasionalID},
		{ID: jabarID, Name: "Jawa Barat", Slug: "jawa-barat", Type: "provinsi", ParentID: &nasionalID},
		{ID: jatimID, Name: "Jawa Timur", Slug: "jawa-timur", Type: "provinsi", ParentID: &nasionalID},
		// Level 3: Kota
		{ID: uuid.New(), Name: "Jakarta Selatan", Slug: "jakarta-selatan", Type: "kota", ParentID: &dkiID},
		{ID: uuid.New(), Name: "Bandung", Slug: "bandung", Type: "kota", ParentID: &jabarID},
		{ID: uuid.New(), Name: "Surabaya", Slug: "surabaya", Type: "kota", ParentID: &jatimID},
	}

	for _, r := range regions {
		if err := db.Create(&r).Error; err != nil {
			log.Printf("⚠️  Failed to seed region '%s': %v", r.Name, err)
		}
	}
	log.Printf("✅ %d default regions seeded", len(regions))
}

// SeedSampleData creates 3 sample users + 3 approved posts + comments + votes for development.
// Only runs if no posts exist yet.
func SeedSampleData(db *gorm.DB) {
	var postCount int64
	db.Model(&postModels.Post{}).Count(&postCount)
	if postCount > 0 {
		return
	}

	// ── 1. Lookup sectors & regions that were seeded earlier ──
	var sectorPendidikan, sectorKesehatan, sectorTransportasi sectorModels.Sector
	db.Where("slug = ?", "pendidikan").First(&sectorPendidikan)
	db.Where("slug = ?", "kesehatan").First(&sectorKesehatan)
	db.Where("slug = ?", "transportasi").First(&sectorTransportasi)

	var regionDKI, regionJabar, regionNasional regionModels.Region
	db.Where("slug = ?", "dki-jakarta").First(&regionDKI)
	db.Where("slug = ?", "jawa-barat").First(&regionJabar)
	db.Where("slug = ?", "nasional").First(&regionNasional)

	// ── 2. Look up admin user (reviewer) ──
	var admin userModels.User
	db.Where("role = ?", helpers.RoleAdministrator).First(&admin)

	// ── 3. Create 3 sample creator users ──
	user1 := userModels.User{
		ID:         uuid.New(),
		Name:       "Rina Wulandari",
		Email:      "rina@example.com",
		AvatarURL:  "https://api.dicebear.com/9.x/avataaars/svg?seed=Rina",
		Bio:        "Pegiat pendidikan di Bandung",
		Location:   "Bandung, Jawa Barat",
		Role:       helpers.RoleCreator,
		Provider:   helpers.ProviderGoogle,
		ProviderID: "seed-google-001",
	}
	user2 := userModels.User{
		ID:         uuid.New(),
		Name:       "Budi Santoso",
		Email:      "budi@example.com",
		AvatarURL:  "https://api.dicebear.com/9.x/avataaars/svg?seed=Budi",
		Bio:        "Dokter PTT, fokus isu kesehatan daerah",
		Location:   "Jakarta Selatan",
		Role:       helpers.RoleCreator,
		Provider:   helpers.ProviderGoogle,
		ProviderID: "seed-google-002",
	}
	user3 := userModels.User{
		ID:         uuid.New(),
		Name:       "Siti Aminah",
		Email:      "siti@example.com",
		AvatarURL:  "https://api.dicebear.com/9.x/avataaars/svg?seed=Siti",
		Bio:        "Komuter harian, peduli transportasi publik",
		Location:   "Depok, Jawa Barat",
		Role:       helpers.RoleCreator,
		Provider:   helpers.ProviderGoogle,
		ProviderID: "seed-google-003",
	}

	for _, u := range []*userModels.User{&user1, &user2, &user3} {
		if err := db.Create(u).Error; err != nil {
			log.Printf("⚠️  Failed to seed user '%s': %v", u.Name, err)
			return
		}
	}
	log.Println("✅ 3 sample users seeded")

	// ── 4. Create 3 approved posts ──
	now := time.Now()

	post1 := postModels.Post{
		ID:       uuid.New(),
		UserID:   user1.ID,
		Title:    "Kurikulum Merdeka Perlu Evaluasi Menyeluruh di Daerah 3T",
		SectorID: &sectorPendidikan.ID,
		RegionID: &regionNasional.ID,
		Criticism: "Implementasi Kurikulum Merdeka di daerah 3T (Terdepan, Terluar, Tertinggal) menghadapi kendala serius. " +
			"Banyak guru belum mendapat pelatihan memadai, infrastruktur digital minim, dan buku pendamping belum terdistribusi merata. " +
			"Akibatnya, gap kualitas pendidikan antara kota besar dan daerah semakin melebar.",
		Solution: "Pemerintah perlu membentuk satuan tugas khusus implementasi kurikulum di daerah 3T dengan 3 langkah: " +
			"(1) Pelatihan guru secara hybrid — online + kunjungan langsung per triwulan, " +
			"(2) Distribusi modul cetak sebagai alternatif digital, " +
			"(3) Kerjasama dengan universitas lokal untuk pendampingan berkelanjutan.",
		ImpactEstimate: "Jika diterapkan, diperkirakan 12.000+ guru di 514 kabupaten/kota 3T bisa mendapat pelatihan dalam 1 tahun pertama.",
		References:     "https://kemdikbud.go.id/kurikulum-merdeka | https://bps.go.id/indikator-pendidikan-2025",
		Status:         helpers.StatusApproved,
		ReviewedBy:     &admin.ID,
		ReviewNote:     "Post berkualitas, data lengkap.",
		VoteCount:      42,
		CommentCount:   3,
		CreatedAt:      now.Add(-72 * time.Hour),
		UpdatedAt:      now.Add(-70 * time.Hour),
	}

	post2 := postModels.Post{
		ID:       uuid.New(),
		UserID:   user2.ID,
		Title:    "Anggaran Puskesmas 2026 Tidak Sebanding dengan Beban Layanan",
		SectorID: &sectorKesehatan.ID,
		RegionID: &regionDKI.ID,
		Criticism: "Puskesmas di DKI Jakarta mengalami lonjakan kunjungan pasien hingga 35% pasca-pandemi, " +
			"namun anggaran operasional 2026 hanya naik 5%. Banyak puskesmas kekurangan obat esensial, " +
			"alat lab aus, dan tenaga medis burnout akibat beban kerja berlebih.",
		Solution: "Alokasi dana khusus 'Puskesmas Recovery Fund' sebesar 15% dari total anggaran kesehatan DKI. " +
			"Dana dialokasikan untuk: pengadaan obat esensial, pembaruan alat diagnostik dasar, " +
			"dan insentif tambahan untuk tenaga medis yang menangani lebih dari 50 pasien/hari.",
		ImpactEstimate: "340 puskesmas di DKI Jakarta akan terdampak positif, melayani 8.5 juta penduduk.",
		References:     "https://dinkes.jakarta.go.id/laporan-2025 | https://kemenkes.go.id/puskesmas-data",
		Status:         helpers.StatusApproved,
		ReviewedBy:     &admin.ID,
		ReviewNote:     "Kritik valid, solusi bisa ditindaklanjuti.",
		VoteCount:      87,
		CommentCount:   2,
		CreatedAt:      now.Add(-48 * time.Hour),
		UpdatedAt:      now.Add(-46 * time.Hour),
	}

	post3 := postModels.Post{
		ID:       uuid.New(),
		UserID:   user3.ID,
		Title:    "Integrasi Tiket Transportasi Jabodetabek Masih Jalan di Tempat",
		SectorID: &sectorTransportasi.ID,
		RegionID: &regionJabar.ID,
		Criticism: "Sejak 2023, pemerintah menjanjikan satu kartu untuk semua moda transportasi Jabodetabek. " +
			"Kenyataannya di 2026, pengguna masih harus tap kartu berbeda untuk KRL, TransJakarta, MRT, dan LRT. " +
			"Sistem top-up tidak interoperable, dan data perjalanan tidak terintegrasi.",
		Solution: "Percepat implementasi JBSC (Jabodetabek Smart Card) dengan target 6 bulan: " +
			"(1) Wajibkan semua operator transit mengadopsi standar NFC tunggal, " +
			"(2) Buat API terbuka untuk saldo lintas operator, " +
			"(3) Sediakan periode transisi 3 bulan di mana kartu lama masih bisa dipakai.",
		ImpactEstimate: "15 juta pengguna transportasi publik harian akan mendapat manfaat langsung.",
		References:     "https://bptj.go.id/integrasi-transportasi | https://jaklingko.id/roadmap",
		Status:         helpers.StatusApproved,
		ReviewedBy:     &admin.ID,
		ReviewNote:     "Topik penting, referensi lengkap.",
		VoteCount:      156,
		CommentCount:   4,
		CreatedAt:      now.Add(-24 * time.Hour),
		UpdatedAt:      now.Add(-22 * time.Hour),
	}

	for _, p := range []*postModels.Post{&post1, &post2, &post3} {
		if err := db.Create(p).Error; err != nil {
			log.Printf("⚠️  Failed to seed post '%s': %v", p.Title, err)
			return
		}
	}
	log.Println("✅ 3 sample posts seeded")

	// ── 5. Create sample votes ──
	votes := []voteModels.Vote{
		// Post 1 votes
		{ID: uuid.New(), UserID: user2.ID, PostID: post1.ID, Type: helpers.VoteUp},
		{ID: uuid.New(), UserID: user3.ID, PostID: post1.ID, Type: helpers.VoteUp},
		// Post 2 votes
		{ID: uuid.New(), UserID: user1.ID, PostID: post2.ID, Type: helpers.VoteUp},
		{ID: uuid.New(), UserID: user3.ID, PostID: post2.ID, Type: helpers.VoteUp},
		// Post 3 votes
		{ID: uuid.New(), UserID: user1.ID, PostID: post3.ID, Type: helpers.VoteUp},
		{ID: uuid.New(), UserID: user2.ID, PostID: post3.ID, Type: helpers.VoteUp},
	}

	for _, v := range votes {
		if err := db.Create(&v).Error; err != nil {
			log.Printf("⚠️  Failed to seed vote: %v", err)
		}
	}
	log.Println("✅ Sample votes seeded")

	// ── 6. Create sample comments (with 1-level replies) ──
	// Post 1 comments
	c1 := commentModels.Comment{
		ID:        uuid.New(),
		UserID:    user2.ID,
		PostID:    post1.ID,
		Content:   "Setuju sekali. Sebagai tenaga kesehatan, saya melihat langsung betapa sulitnya guru di daerah terpencil mengakses materi digital.",
		CreatedAt: now.Add(-68 * time.Hour),
	}
	c1Reply := commentModels.Comment{
		ID:        uuid.New(),
		UserID:    user1.ID,
		PostID:    post1.ID,
		ParentID:  &c1.ID,
		Content:   "Terima kasih Budi! Memang perlu pendekatan non-digital juga. Modul cetak bisa jadi solusi antara.",
		CreatedAt: now.Add(-66 * time.Hour),
	}
	c1b := commentModels.Comment{
		ID:        uuid.New(),
		UserID:    user3.ID,
		PostID:    post1.ID,
		Content:   "Apakah ada data berapa persen guru 3T yang sudah punya akses internet stabil?",
		CreatedAt: now.Add(-64 * time.Hour),
	}

	// Post 2 comments
	c2 := commentModels.Comment{
		ID:        uuid.New(),
		UserID:    user3.ID,
		PostID:    post2.ID,
		Content:   "Puskesmas dekat rumah saya di Depok juga mengalami hal serupa. Antrian panjang, obat sering kosong.",
		CreatedAt: now.Add(-44 * time.Hour),
	}
	c2Reply := commentModels.Comment{
		ID:        uuid.New(),
		UserID:    user2.ID,
		PostID:    post2.ID,
		ParentID:  &c2.ID,
		Content:   "Memang ini masalah sistemik. Semoga pemerintah daerah bisa mengalokasikan anggaran darurat.",
		CreatedAt: now.Add(-42 * time.Hour),
	}

	// Post 3 comments
	c3 := commentModels.Comment{
		ID:        uuid.New(),
		UserID:    user1.ID,
		PostID:    post3.ID,
		Content:   "Sebagai pengguna KRL harian, ini sangat mengganggu. Harus bawa 3 kartu setiap hari.",
		CreatedAt: now.Add(-20 * time.Hour),
	}
	c3Reply := commentModels.Comment{
		ID:        uuid.New(),
		UserID:    user3.ID,
		PostID:    post3.ID,
		ParentID:  &c3.ID,
		Content:   "Sama! Belum lagi kalau salah satu kartu saldo habis, harus antri top-up terpisah.",
		CreatedAt: now.Add(-18 * time.Hour),
	}
	c3b := commentModels.Comment{
		ID:        uuid.New(),
		UserID:    user2.ID,
		PostID:    post3.ID,
		Content:   "Di banyak kota besar di Asia seperti Seoul dan Singapura, integrasi tiket sudah berjalan puluhan tahun. Kita tertinggal.",
		CreatedAt: now.Add(-16 * time.Hour),
	}
	c3bReply := commentModels.Comment{
		ID:        uuid.New(),
		UserID:    user3.ID,
		PostID:    post3.ID,
		ParentID:  &c3b.ID,
		Content:   "Betul, bahkan Bangkok juga sudah mulai integrasi. Mestinya jadi benchmark.",
		CreatedAt: now.Add(-14 * time.Hour),
	}

	comments := []*commentModels.Comment{
		&c1, &c1Reply, &c1b,
		&c2, &c2Reply,
		&c3, &c3Reply, &c3b, &c3bReply,
	}

	for _, c := range comments {
		if err := db.Create(c).Error; err != nil {
			log.Printf("⚠️  Failed to seed comment: %v", err)
		}
	}
	log.Printf("✅ %d sample comments seeded", len(comments))

	log.Println("🎉 All sample data seeded successfully!")
}
