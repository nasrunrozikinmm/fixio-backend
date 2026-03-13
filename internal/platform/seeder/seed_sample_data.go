package seeder

import (
	"log"
	"time"

	userModels "fixio/internal/modules/auth/entity"
	bookmarkModels "fixio/internal/modules/bookmark/entity"
	commentModels "fixio/internal/modules/comment/entity"
	followModels "fixio/internal/modules/follow/entity"
	notifModels "fixio/internal/modules/notification/entity"
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

// SeedSampleData creates comprehensive sample data for development:
// - 4 users (+ existing admin)
// - 6 approved, 4 pending_review, 1 draft, 2 rejected posts
// - Votes, comments with replies, bookmarks, follows, notifications
// Only runs if no posts exist yet.
func SeedSampleData(db *gorm.DB) {
	var postCount int64
	db.Model(&postModels.Post{}).Count(&postCount)
	if postCount > 0 {
		return
	}

	// ── 1. Lookup sectors & regions that were seeded earlier ──
	var sectorPendidikan, sectorKesehatan, sectorTransportasi sectorModels.Sector
	var sectorEkonomi, sectorLingkungan, sectorTeknologi sectorModels.Sector
	db.Where("slug = ?", "pendidikan").First(&sectorPendidikan)
	db.Where("slug = ?", "kesehatan").First(&sectorKesehatan)
	db.Where("slug = ?", "transportasi").First(&sectorTransportasi)
	db.Where("slug = ?", "ekonomi").First(&sectorEkonomi)
	db.Where("slug = ?", "lingkungan").First(&sectorLingkungan)
	db.Where("slug = ?", "teknologi").First(&sectorTeknologi)

	var regionDKI, regionJabar, regionNasional, regionJatim regionModels.Region
	db.Where("slug = ?", "dki-jakarta").First(&regionDKI)
	db.Where("slug = ?", "jawa-barat").First(&regionJabar)
	db.Where("slug = ?", "nasional").First(&regionNasional)
	db.Where("slug = ?", "jawa-timur").First(&regionJatim)

	// ── 2. Look up admin user (reviewer) ──
	var admin userModels.User
	db.Where("role = ?", helpers.RoleAdministrator).First(&admin)

	// ── 3. Create 4 sample users (1 moderator + 3 creators) ──
	moderator := userModels.User{
		ID:         uuid.New(),
		Name:       "Dewi Kartini",
		Email:      "dewi@example.com",
		AvatarURL:  "https://api.dicebear.com/9.x/avataaars/svg?seed=Dewi",
		Bio:        "Moderator komunitas, mantan jurnalis investigasi",
		Location:   "Yogyakarta",
		Role:       helpers.RoleModerator,
		Provider:   helpers.ProviderGoogle,
		ProviderID: "seed-google-000",
	}

	user1 := userModels.User{
		ID:         uuid.New(),
		Name:       "Rina Wulandari",
		Email:      "rina@example.com",
		AvatarURL:  "https://api.dicebear.com/9.x/avataaars/svg?seed=Rina",
		Bio:        "Pegiat pendidikan di Bandung. Guru honorer selama 8 tahun.",
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
		Bio:        "Dokter PTT di puskesmas, fokus isu kesehatan daerah terpencil",
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
		Bio:        "Komuter harian Depok-Jakarta, advokat transportasi publik",
		Location:   "Depok, Jawa Barat",
		Role:       helpers.RoleCreator,
		Provider:   helpers.ProviderGoogle,
		ProviderID: "seed-google-003",
	}

	for _, u := range []*userModels.User{&moderator, &user1, &user2, &user3} {
		if err := db.Create(u).Error; err != nil {
			log.Printf("⚠️  Failed to seed user '%s': %v", u.Name, err)
			return
		}
	}
	log.Println("✅ 4 sample users seeded (1 moderator + 3 creators)")

	// ── 4. Create 13 posts (6 approved, 4 pending_review, 1 draft, 2 rejected) ──
	now := time.Now()

	post1 := postModels.Post{
		ID:       uuid.New(),
		UserID:   user1.ID,
		Title:    "Kurikulum Merdeka Perlu Evaluasi Menyeluruh di Daerah 3T",
		SectorID: &sectorPendidikan.ID,
		RegionID: &regionNasional.ID,
		Criticism: "<p>Implementasi <strong>Kurikulum Merdeka</strong> di daerah 3T (Terdepan, Terluar, Tertinggal) menghadapi kendala serius.</p>" +
			"<p>Banyak guru belum mendapat pelatihan memadai, infrastruktur digital minim, dan buku pendamping belum terdistribusi merata. " +
			"Akibatnya, <em>gap kualitas pendidikan</em> antara kota besar dan daerah semakin melebar.</p>" +
			"<p>Data BPS 2025 menunjukkan hanya 23% sekolah di daerah 3T yang memiliki akses internet stabil.</p>",
		Solution: "<p>Pemerintah perlu membentuk satuan tugas khusus implementasi kurikulum di daerah 3T dengan 3 langkah:</p>" +
			"<ol><li>Pelatihan guru secara hybrid — online + kunjungan langsung per triwulan</li>" +
			"<li>Distribusi modul cetak sebagai alternatif digital</li>" +
			"<li>Kerjasama dengan universitas lokal untuk pendampingan berkelanjutan</li></ol>",
		ImpactEstimate: "Jika diterapkan, diperkirakan 12.000+ guru di 514 kabupaten/kota 3T bisa mendapat pelatihan dalam 1 tahun pertama.",
		References:     "https://kemdikbud.go.id/kurikulum-merdeka | https://bps.go.id/indikator-pendidikan-2025",
		Images: postModels.StringArray{
			"https://images.unsplash.com/photo-1503676260728-1c00da094a0b?w=800",
			"https://images.unsplash.com/photo-1509062522246-3755977927d7?w=800",
		},
		Status:       helpers.StatusApproved,
		ReviewedBy:   &admin.ID,
		ReviewNote:   "Post berkualitas, data lengkap dan referensi valid.",
		VoteCount:    42,
		CommentCount: 3,
		CreatedAt:    now.Add(-168 * time.Hour),
		UpdatedAt:    now.Add(-166 * time.Hour),
	}

	post2 := postModels.Post{
		ID:       uuid.New(),
		UserID:   user2.ID,
		Title:    "Anggaran Puskesmas 2026 Tidak Sebanding dengan Beban Layanan",
		SectorID: &sectorKesehatan.ID,
		RegionID: &regionDKI.ID,
		Criticism: "<p>Puskesmas di DKI Jakarta mengalami <strong>lonjakan kunjungan pasien hingga 35%</strong> pasca-pandemi, " +
			"namun anggaran operasional 2026 hanya naik 5%.</p>" +
			"<p>Banyak puskesmas kekurangan obat esensial, alat lab aus, dan tenaga medis <em>burnout</em> akibat beban kerja berlebih. " +
			"Rata-rata satu dokter di puskesmas DKI menangani 60-80 pasien per hari.</p>",
		Solution: "<p>Alokasi dana khusus <strong>'Puskesmas Recovery Fund'</strong> sebesar 15% dari total anggaran kesehatan DKI:</p>" +
			"<ul><li>Pengadaan obat esensial dan vaksin</li>" +
			"<li>Pembaruan alat diagnostik dasar (USG, lab kit)</li>" +
			"<li>Insentif tambahan untuk tenaga medis yang menangani >50 pasien/hari</li></ul>",
		ImpactEstimate: "340 puskesmas di DKI Jakarta akan terdampak positif, melayani 8.5 juta penduduk.",
		References:     "https://dinkes.jakarta.go.id/laporan-2025 | https://kemenkes.go.id/puskesmas-data",
		Images: postModels.StringArray{
			"https://images.unsplash.com/photo-1519494026892-80bbd2d6fd0d?w=800",
			"https://images.unsplash.com/photo-1631815588090-d4bfec5b1b89?w=800",
			"https://images.unsplash.com/photo-1584982751601-97dcc096659c?w=800",
		},
		Status:       helpers.StatusApproved,
		ReviewedBy:   &admin.ID,
		ReviewNote:   "Kritik valid dengan data kuantitatif, solusi bisa ditindaklanjuti.",
		VoteCount:    87,
		CommentCount: 2,
		CreatedAt:    now.Add(-120 * time.Hour),
		UpdatedAt:    now.Add(-118 * time.Hour),
	}

	post3 := postModels.Post{
		ID:       uuid.New(),
		UserID:   user3.ID,
		Title:    "Integrasi Tiket Transportasi Jabodetabek Masih Jalan di Tempat",
		SectorID: &sectorTransportasi.ID,
		RegionID: &regionJabar.ID,
		Criticism: "<p>Sejak 2023, pemerintah menjanjikan <strong>satu kartu untuk semua moda transportasi</strong> Jabodetabek.</p>" +
			"<p>Kenyataannya di 2026, pengguna masih harus tap kartu berbeda untuk KRL, TransJakarta, MRT, dan LRT. " +
			"Sistem top-up tidak interoperable, dan data perjalanan tidak terintegrasi.</p>" +
			"<p>Ini menyebabkan inefisiensi waktu rata-rata 15 menit per hari per komuter.</p>",
		Solution: "<p>Percepat implementasi <strong>JBSC (Jabodetabek Smart Card)</strong> dengan target 6 bulan:</p>" +
			"<ol><li>Wajibkan semua operator transit mengadopsi standar NFC tunggal</li>" +
			"<li>Buat API terbuka untuk saldo lintas operator</li>" +
			"<li>Sediakan periode transisi 3 bulan di mana kartu lama masih bisa dipakai</li></ol>",
		ImpactEstimate: "15 juta pengguna transportasi publik harian akan mendapat manfaat langsung.",
		References:     "https://bptj.go.id/integrasi-transportasi | https://jaklingko.id/roadmap",
		Images: postModels.StringArray{
			"https://images.unsplash.com/photo-1544620347-c4fd4a3d5957?w=800",
		},
		Status:       helpers.StatusApproved,
		ReviewedBy:   &admin.ID,
		ReviewNote:   "Topik penting, referensi lengkap.",
		VoteCount:    156,
		CommentCount: 4,
		CreatedAt:    now.Add(-96 * time.Hour),
		UpdatedAt:    now.Add(-94 * time.Hour),
	}

	post4 := postModels.Post{
		ID:       uuid.New(),
		UserID:   user1.ID,
		Title:    "Sampah Plastik di Sungai Citarum: Kebijakan Daerah Tidak Efektif",
		SectorID: &sectorLingkungan.ID,
		RegionID: &regionJabar.ID,
		Criticism: "<p>Program <strong>Citarum Harum</strong> yang diluncurkan sejak 2018 belum menunjukkan hasil signifikan.</p>" +
			"<p>Volume sampah plastik di Sungai Citarum masih mencapai 20.000 ton/tahun. " +
			"Regulasi daerah tentang larangan plastik sekali pakai tidak diimplementasikan secara konsisten, " +
			"dan penegakan hukum terhadap industri pembuang limbah masih <em>sangat lemah</em>.</p>",
		Solution: "<p>Pendekatan tiga pilar untuk Sungai Citarum:</p>" +
			"<ol><li><strong>Hulu:</strong> Subsidi 50% untuk UMKM yang beralih ke kemasan non-plastik</li>" +
			"<li><strong>Tengah:</strong> Pasang 20 trash boom di titik kritis, dikelola oleh BUMDes</li>" +
			"<li><strong>Hilir:</strong> Bangun 5 fasilitas daur ulang skala komunitas dengan pembiayaan CSR</li></ol>",
		ImpactEstimate: "Potensi pengurangan sampah plastik 40% dalam 2 tahun di sepanjang aliran Citarum.",
		References:     "https://dlh.jabarprov.go.id/citarum-harum | https://walhi.or.id/citarum-report-2025",
		Images: postModels.StringArray{
			"https://images.unsplash.com/photo-1621451537084-482c73073a0f?w=800",
			"https://images.unsplash.com/photo-1532996122724-e3c354a0b15b?w=800",
			"https://images.unsplash.com/photo-1504711434969-e33886168d6c?w=800",
			"https://images.unsplash.com/photo-1611284446314-60a58ac0deb9?w=800",
		},
		Status:       helpers.StatusApproved,
		ReviewedBy:   &moderator.ID,
		ReviewNote:   "Data kuat, solusi terstruktur dan realistis.",
		VoteCount:    63,
		CommentCount: 2,
		CreatedAt:    now.Add(-72 * time.Hour),
		UpdatedAt:    now.Add(-70 * time.Hour),
	}

	post5 := postModels.Post{
		ID:       uuid.New(),
		UserID:   user2.ID,
		Title:    "UMKM Digital: Pelatihan Pemerintah Tidak Menyentuh Akar Masalah",
		SectorID: &sectorEkonomi.ID,
		RegionID: &regionJatim.ID,
		Criticism: "<p>Program pelatihan digitalisasi UMKM dari Kemenparekraf berfokus pada pembuatan akun marketplace, " +
			"padahal <strong>80% UMKM</strong> yang gagal go-digital terkendala di hal yang lebih mendasar:</p>" +
			"<ul><li>Literasi keuangan rendah — tidak bisa menghitung margin</li>" +
			"<li>Koneksi internet tidak stabil di area pedesaan</li>" +
			"<li>Tidak ada pendampingan pasca-pelatihan</li></ul>",
		Solution: "<p>Redesain program pelatihan UMKM digital dengan pendekatan <strong>berjenjang</strong>:</p>" +
			"<ol><li><strong>Level 1 (Fondasi):</strong> Literasi keuangan + pencatatan digital sederhana (3 bulan)</li>" +
			"<li><strong>Level 2 (Go-Online):</strong> Setup marketplace + konten produk (3 bulan, setelah lulus L1)</li>" +
			"<li><strong>Level 3 (Scale):</strong> Iklan digital + analitik penjualan (3 bulan, setelah lulus L2)</li></ol>" +
			"<p>Setiap level disertai <em>mentor pendamping</em> dari mahasiswa KKN digital.</p>",
		ImpactEstimate: "Jika diterapkan di 10 kota pilot, bisa menjangkau 50.000 UMKM dalam tahun pertama.",
		References:     "https://kemenparekraf.go.id/umkm-digital-2025 | https://bps.go.id/umkm-data",
		Images: postModels.StringArray{
			"https://images.unsplash.com/photo-1556740758-90de374c12ad?w=800",
			"https://images.unsplash.com/photo-1542744173-8e7e53415bb0?w=800",
		},
		Status:       helpers.StatusApproved,
		ReviewedBy:   &moderator.ID,
		ReviewNote:   "Analisis mendalam, solusi bertahap yang realistis.",
		VoteCount:    34,
		CommentCount: 1,
		CreatedAt:    now.Add(-48 * time.Hour),
		UpdatedAt:    now.Add(-46 * time.Hour),
	}

	post6 := postModels.Post{
		ID:       uuid.New(),
		UserID:   user3.ID,
		Title:    "Keamanan Data Pribadi di Aplikasi Pemerintah Masih Rentan",
		SectorID: &sectorTeknologi.ID,
		RegionID: &regionNasional.ID,
		Criticism: "<p>Dalam 2 tahun terakhir, setidaknya <strong>5 aplikasi pemerintah</strong> mengalami insiden kebocoran data:</p>" +
			"<ul><li>Aplikasi PeduliLindungi — 1.3 juta data terekspos (2024)</li>" +
			"<li>SiPetani — 500 ribu data petani bocor (2025)</li>" +
			"<li>DTKS Kemensos — data penerima bantuan tersebar di forum gelap</li></ul>" +
			"<p>Audit keamanan tidak wajib sebelum launch, dan tidak ada <em>consequences</em> jika terjadi kebocoran.</p>",
		Solution: "<p>Tiga kebijakan wajib untuk semua aplikasi pemerintah:</p>" +
			"<ol><li><strong>Mandatory Security Audit</strong> — setiap aplikasi wajib lulus penetration testing sebelum rilis</li>" +
			"<li><strong>Data Protection Officer (DPO)</strong> — setiap K/L yang mengelola data warga wajib punya DPO bersertifikat</li>" +
			"<li><strong>Sanksi Administratif</strong> — pejabat TI bertanggung jawab langsung jika terjadi kebocoran</li></ol>",
		ImpactEstimate: "270 juta data warga Indonesia akan lebih terlindungi.",
		References:     "https://kominfo.go.id/kebocoran-data-2025 | https://bssn.go.id/audit-keamanan",
		Images: postModels.StringArray{
			"https://images.unsplash.com/photo-1550751827-4bd374c3f58b?w=800",
			"https://images.unsplash.com/photo-1563986768609-322da13575f2?w=800",
			"https://images.unsplash.com/photo-1555949963-ff9fe0c870eb?w=800",
			"https://images.unsplash.com/photo-1526374965328-7f61d4dc18c5?w=800",
			"https://images.unsplash.com/photo-1510511459019-5dda7724fd87?w=800",
		},
		Status:       helpers.StatusApproved,
		ReviewedBy:   &admin.ID,
		ReviewNote:   "Isu kritis, data insiden valid dan terdokumentasi.",
		VoteCount:    211,
		CommentCount: 3,
		CreatedAt:    now.Add(-24 * time.Hour),
		UpdatedAt:    now.Add(-22 * time.Hour),
	}

	// Pending review post (no images yet — user hasn't added)
	post7 := postModels.Post{
		ID:       uuid.New(),
		UserID:   user1.ID,
		Title:    "Subsidi Pupuk 2026 Terlambat Sampai ke Petani Kecil",
		SectorID: &sectorEkonomi.ID,
		RegionID: &regionJatim.ID,
		Criticism: "<p>Subsidi pupuk yang dianggarkan untuk musim tanam Maret 2026 baru sampai ke petani di Jawa Timur pada " +
			"bulan Mei — <strong>2 bulan terlambat</strong>. Petani terpaksa membeli pupuk non-subsidi dengan harga 3x lipat.</p>" +
			"<p>Masalah utama: distribusi melalui 3 lapisan birokrasi (pusat → provinsi → kabupaten → distributor).</p>",
		Solution: "<p>Reformasi rantai distribusi subsidi pupuk:</p>" +
			"<ol><li>Pangkas lapisan distribusi dari 4 menjadi 2 (pusat → koperasi tani)</li>" +
			"<li>Gunakan e-voucher berbasis NIK agar petani bisa tebus di toko manapun</li>" +
			"<li>Dashboard real-time stok pupuk per kabupaten, terbuka untuk publik</li></ol>",
		ImpactEstimate: "4.2 juta petani kecil di Jawa Timur bisa mendapat pupuk tepat waktu.",
		References:     "https://kementan.go.id/subsidi-pupuk-2026",
		Images:         postModels.StringArray{},
		Status:         helpers.StatusPendingReview,
		VoteCount:      0,
		CommentCount:   0,
		CreatedAt:      now.Add(-6 * time.Hour),
		UpdatedAt:      now.Add(-6 * time.Hour),
	}

	// Draft post
	post8 := postModels.Post{
		ID:        uuid.New(),
		UserID:    user2.ID,
		Title:     "Kekurangan Dokter Spesialis di RS Daerah",
		SectorID:  &sectorKesehatan.ID,
		RegionID:  &regionJatim.ID,
		Criticism: "<p>RS daerah di Jawa Timur rata-rata hanya memiliki 2-3 dokter spesialis, jauh di bawah standar WHO.</p>",
		Solution:  "<p>(masih dalam penyusunan)</p>",
		Images:    postModels.StringArray{},
		Status:    helpers.StatusDraft,
		VoteCount: 0,
		CreatedAt: now.Add(-2 * time.Hour),
		UpdatedAt: now.Add(-2 * time.Hour),
	}

	// ── Additional pending_review posts (for Antrian Review page) ──

	post9 := postModels.Post{
		ID:       uuid.New(),
		UserID:   user2.ID,
		Title:    "Akses Air Bersih di Nusa Tenggara Timur Masih Jauh dari Target SDGs",
		SectorID: &sectorKesehatan.ID,
		RegionID: &regionNasional.ID,
		Criticism: "<p>Target SDGs 2030 mensyaratkan <strong>100% akses air bersih</strong>, namun di NTT " +
			"hanya 42% rumah tangga yang memiliki akses air layak minum.</p>" +
			"<p>Program PAMSIMAS yang dibiayai pemerintah pusat " +
			"sering mangkrak karena kurangnya tenaga teknis pemeliharaan di tingkat desa.</p>",
		Solution: "<p>Strategi tiga lapis:</p>" +
			"<ol><li>Rekrut dan latih 500 teknisi air desa dari lulusan SMK setempat</li>" +
			"<li>Gunakan teknologi pompa tenaga surya untuk daerah tanpa listrik</li>" +
			"<li>Alokasikan dana abadi desa untuk biaya pemeliharaan tahunan</li></ol>",
		ImpactEstimate: "1.2 juta warga NTT bisa mendapat akses air bersih dalam 3 tahun.",
		References:     "https://pamsimas.pu.go.id | https://bps.go.id/air-bersih-ntt",
		Images: postModels.StringArray{
			"https://images.unsplash.com/photo-1541544741938-0af808871cc0?w=800",
		},
		Status:       helpers.StatusPendingReview,
		VoteCount:    0,
		CommentCount: 0,
		CreatedAt:    now.Add(-5 * time.Hour),
		UpdatedAt:    now.Add(-5 * time.Hour),
	}

	post10 := postModels.Post{
		ID:       uuid.New(),
		UserID:   user3.ID,
		Title:    "Kemacetan Bandung Makin Parah, BRT Tidak Kunjung Beroperasi",
		SectorID: &sectorTransportasi.ID,
		RegionID: &regionJabar.ID,
		Criticism: "<p>Proyek <strong>BRT (Bus Rapid Transit) Bandung Raya</strong> yang dijanjikan sejak 2022 " +
			"belum juga beroperasi di 2026. Sementara itu, jumlah kendaraan pribadi naik 8% per tahun.</p>" +
			"<p>Waktu tempuh rata-rata dari Cimahi ke pusat Bandung sudah mencapai 90 menit pada jam sibuk.</p>",
		Solution: "<p>Percepat operasional BRT koridor 1 (Cimahi-Leuwipanjang-Cicaheum):</p>" +
			"<ol><li>Operasikan jalur darurat menggunakan bus existing sambil menunggu infrastruktur permanen</li>" +
			"<li>Terapkan congestion pricing di kawasan Dago dan Braga pada jam sibuk</li>" +
			"<li>Integrasikan jadwal BRT dengan kereta api lokal Bandung Raya</li></ol>",
		ImpactEstimate: "Potensi pengurangan kemacetan 25% di koridor utama Bandung.",
		References:     "https://dishub.bandung.go.id/brt-2026 | https://bps.go.id/transportasi-jabar",
		Images:         postModels.StringArray{},
		Status:         helpers.StatusPendingReview,
		VoteCount:      0,
		CommentCount:   0,
		CreatedAt:      now.Add(-3 * time.Hour),
		UpdatedAt:      now.Add(-3 * time.Hour),
	}

	post11 := postModels.Post{
		ID:       uuid.New(),
		UserID:   user1.ID,
		Title:    "Dana Desa untuk Pendidikan Sering Dialihkan ke Infrastruktur Fisik",
		SectorID: &sectorPendidikan.ID,
		RegionID: &regionJatim.ID,
		Criticism: "<p>Banyak kepala desa di Jawa Timur mengalokasikan <strong>80% dana desa</strong> untuk pembangunan jalan " +
			"dan gedung, sementara program beasiswa dan pelatihan guru hanya mendapat <em>kurang dari 5%</em>.</p>" +
			"<p>Akibatnya, anak-anak desa terpaksa putus sekolah karena biaya transport ke SMP/SMA di kecamatan.</p>",
		Solution: "<p>Wajibkan alokasi minimum 15% dana desa untuk pendidikan:</p>" +
			"<ol><li>Beasiswa transportasi untuk siswa SMP/SMA jarak jauh</li>" +
			"<li>Honor tambahan untuk guru honorer desa</li>" +
			"<li>Pengadaan perpustakaan keliling per kecamatan</li></ol>",
		ImpactEstimate: "Berpotensi menurunkan angka putus sekolah di desa 30% dalam 2 tahun.",
		References:     "https://kemendesa.go.id/dana-desa-2026",
		Images:         postModels.StringArray{},
		Status:         helpers.StatusPendingReview,
		VoteCount:      0,
		CommentCount:   0,
		CreatedAt:      now.Add(-1 * time.Hour),
		UpdatedAt:      now.Add(-1 * time.Hour),
	}

	// ── Additional rejected posts (for Riwayat / History page) ──

	post12 := postModels.Post{
		ID:           uuid.New(),
		UserID:       user2.ID,
		Title:        "Semua Pejabat Korup Harus Dihukum Mati",
		SectorID:     &sectorLingkungan.ID,
		RegionID:     &regionNasional.ID,
		Criticism:    "<p>Korupsi sudah merajalela di semua lini pemerintahan. Tidak ada yang bisa dipercaya.</p>",
		Solution:     "<p>Hukum mati semua koruptor tanpa pengadilan.</p>",
		Images:       postModels.StringArray{},
		Status:       helpers.StatusRejected,
		ReviewedBy:   &moderator.ID,
		ReviewNote:   "Post tidak memenuhi standar: tidak ada data pendukung, solusi tidak konstruktif, dan mengandung seruan kekerasan.",
		VoteCount:    0,
		CommentCount: 0,
		CreatedAt:    now.Add(-200 * time.Hour),
		UpdatedAt:    now.Add(-198 * time.Hour),
	}

	post13 := postModels.Post{
		ID:       uuid.New(),
		UserID:   user3.ID,
		Title:    "WiFi Gratis di Ruang Publik Jakarta Sering Mati, Anggaran Kemana?",
		SectorID: &sectorTeknologi.ID,
		RegionID: &regionDKI.ID,
		Criticism: "<p>Program <strong>Jakarta Smart City WiFi</strong> menjanjikan 500 titik hotspot gratis " +
			"di ruang publik, tetapi berdasarkan survei mandiri kami, hanya 35% yang benar-benar aktif.</p>",
		Solution:     "<p>Audit titik WiFi secara berkala dan publikasikan dashboard status real-time.</p>",
		Images:       postModels.StringArray{},
		Status:       helpers.StatusRejected,
		ReviewedBy:   &admin.ID,
		ReviewNote:   "Kritik valid namun solusi terlalu singkat. Perlu data survei yang bisa diverifikasi. Silakan perbaiki dan submit ulang.",
		VoteCount:    0,
		CommentCount: 0,
		CreatedAt:    now.Add(-150 * time.Hour),
		UpdatedAt:    now.Add(-148 * time.Hour),
	}

	for _, p := range []*postModels.Post{&post1, &post2, &post3, &post4, &post5, &post6, &post7, &post8, &post9, &post10, &post11, &post12, &post13} {
		if err := db.Create(p).Error; err != nil {
			log.Printf("⚠️  Failed to seed post '%s': %v", p.Title, err)
			return
		}
	}
	log.Println("✅ 13 sample posts seeded (6 approved, 4 pending, 1 draft, 2 rejected)")

	// ── 5. Create sample votes (diverse patterns) ──
	votes := []voteModels.Vote{
		// Post 1 — Kurikulum Merdeka
		{ID: uuid.New(), UserID: user2.ID, PostID: post1.ID, Type: helpers.VoteUp},
		{ID: uuid.New(), UserID: user3.ID, PostID: post1.ID, Type: helpers.VoteUp},
		{ID: uuid.New(), UserID: moderator.ID, PostID: post1.ID, Type: helpers.VoteUp},
		// Post 2 — Puskesmas
		{ID: uuid.New(), UserID: user1.ID, PostID: post2.ID, Type: helpers.VoteUp},
		{ID: uuid.New(), UserID: user3.ID, PostID: post2.ID, Type: helpers.VoteUp},
		{ID: uuid.New(), UserID: moderator.ID, PostID: post2.ID, Type: helpers.VoteUp},
		// Post 3 — Transportasi
		{ID: uuid.New(), UserID: user1.ID, PostID: post3.ID, Type: helpers.VoteUp},
		{ID: uuid.New(), UserID: user2.ID, PostID: post3.ID, Type: helpers.VoteUp},
		{ID: uuid.New(), UserID: moderator.ID, PostID: post3.ID, Type: helpers.VoteUp},
		// Post 4 — Citarum
		{ID: uuid.New(), UserID: user2.ID, PostID: post4.ID, Type: helpers.VoteUp},
		{ID: uuid.New(), UserID: user3.ID, PostID: post4.ID, Type: helpers.VoteUp},
		// Post 5 — UMKM
		{ID: uuid.New(), UserID: user1.ID, PostID: post5.ID, Type: helpers.VoteUp},
		{ID: uuid.New(), UserID: user3.ID, PostID: post5.ID, Type: helpers.VoteDown},
		// Post 6 — Keamanan Data
		{ID: uuid.New(), UserID: user1.ID, PostID: post6.ID, Type: helpers.VoteUp},
		{ID: uuid.New(), UserID: user2.ID, PostID: post6.ID, Type: helpers.VoteUp},
		{ID: uuid.New(), UserID: moderator.ID, PostID: post6.ID, Type: helpers.VoteUp},
	}

	for _, v := range votes {
		if err := db.Create(&v).Error; err != nil {
			log.Printf("⚠️  Failed to seed vote: %v", err)
		}
	}
	log.Printf("✅ %d sample votes seeded", len(votes))

	// ── 6. Create sample comments (with nested replies) ──
	// Post 1 comments
	c1 := commentModels.Comment{
		ID:        uuid.New(),
		UserID:    user2.ID,
		PostID:    post1.ID,
		Content:   "Setuju sekali. Sebagai tenaga kesehatan yang sering ke daerah terpencil, saya melihat langsung betapa sulitnya guru di sana mengakses materi digital.",
		CreatedAt: now.Add(-164 * time.Hour),
	}
	c1Reply := commentModels.Comment{
		ID:        uuid.New(),
		UserID:    user1.ID,
		PostID:    post1.ID,
		ParentID:  &c1.ID,
		Content:   "Terima kasih Budi! Memang perlu pendekatan non-digital juga. Modul cetak bisa jadi solusi antara selama infrastruktur belum merata.",
		CreatedAt: now.Add(-162 * time.Hour),
	}
	c1b := commentModels.Comment{
		ID:        uuid.New(),
		UserID:    user3.ID,
		PostID:    post1.ID,
		Content:   "Apakah ada data berapa persen guru 3T yang sudah punya akses internet stabil? Ini penting untuk menentukan proporsi modul cetak vs digital.",
		CreatedAt: now.Add(-160 * time.Hour),
	}

	// Post 2 comments
	c2 := commentModels.Comment{
		ID:        uuid.New(),
		UserID:    user3.ID,
		PostID:    post2.ID,
		Content:   "Puskesmas dekat rumah saya di Depok juga mengalami hal serupa. Antrian panjang sejak pagi, obat sering kosong. Kasihan tenaga medisnya yang sudah overwork.",
		CreatedAt: now.Add(-116 * time.Hour),
	}
	c2Reply := commentModels.Comment{
		ID:        uuid.New(),
		UserID:    user2.ID,
		PostID:    post2.ID,
		ParentID:  &c2.ID,
		Content:   "Memang ini masalah sistemik. Semoga pemerintah daerah bisa mengalokasikan anggaran darurat. Kami para tenaga medis sudah di titik limit.",
		CreatedAt: now.Add(-114 * time.Hour),
	}

	// Post 3 comments
	c3 := commentModels.Comment{
		ID:        uuid.New(),
		UserID:    user1.ID,
		PostID:    post3.ID,
		Content:   "Sebagai pengguna KRL harian Bandung-Jakarta, ini sangat mengganggu. Harus bawa 3 kartu berbeda setiap hari ke kantor.",
		CreatedAt: now.Add(-92 * time.Hour),
	}
	c3Reply := commentModels.Comment{
		ID:        uuid.New(),
		UserID:    user3.ID,
		PostID:    post3.ID,
		ParentID:  &c3.ID,
		Content:   "Sama! Belum lagi kalau salah satu kartu saldo habis, harus antri top-up terpisah di loket yang beda-beda. Buang waktu 20 menit.",
		CreatedAt: now.Add(-90 * time.Hour),
	}
	c3b := commentModels.Comment{
		ID:        uuid.New(),
		UserID:    user2.ID,
		PostID:    post3.ID,
		Content:   "Di banyak kota besar di Asia seperti Seoul dan Singapura, integrasi tiket sudah berjalan puluhan tahun. Kita sangat tertinggal.",
		CreatedAt: now.Add(-88 * time.Hour),
	}
	c3bReply := commentModels.Comment{
		ID:        uuid.New(),
		UserID:    moderator.ID,
		PostID:    post3.ID,
		ParentID:  &c3b.ID,
		Content:   "Betul. Bangkok juga sudah mulai integrasi dengan Rabbit Card. Mestinya jadi benchmark untuk regulator kita.",
		CreatedAt: now.Add(-86 * time.Hour),
	}

	// Post 4 comments
	c4 := commentModels.Comment{
		ID:        uuid.New(),
		UserID:    user3.ID,
		PostID:    post4.ID,
		Content:   "Saya tinggal di dekat Citarum. Baunya sudah tidak tertahankan, terutama musim kemarau. Anak-anak sering sakit kulit.",
		CreatedAt: now.Add(-68 * time.Hour),
	}
	c4Reply := commentModels.Comment{
		ID:        uuid.New(),
		UserID:    user1.ID,
		PostID:    post4.ID,
		ParentID:  &c4.ID,
		Content:   "Ini yang menyedihkan — dampaknya langsung ke kesehatan warga sekitar. Pemerintah tidak boleh hanya retorika tanpa eksekusi.",
		CreatedAt: now.Add(-66 * time.Hour),
	}

	// Post 5 comment
	c5 := commentModels.Comment{
		ID:        uuid.New(),
		UserID:    moderator.ID,
		PostID:    post5.ID,
		Content:   "Analisis yang bagus. Saya pernah wawancara 30 pelaku UMKM di Surabaya — memang mayoritas terkendala literasi keuangan, bukan akses marketplace.",
		CreatedAt: now.Add(-44 * time.Hour),
	}

	// Post 6 comments
	c6 := commentModels.Comment{
		ID:        uuid.New(),
		UserID:    user1.ID,
		PostID:    post6.ID,
		Content:   "Saya pernah menemukan data siswa saya tersebar di internet karena kebocoran Dapodik. Ini sangat mengerikan bagi orang tua murid.",
		CreatedAt: now.Add(-20 * time.Hour),
	}
	c6Reply := commentModels.Comment{
		ID:        uuid.New(),
		UserID:    user3.ID,
		PostID:    post6.ID,
		ParentID:  &c6.ID,
		Content:   "Harusnya ada kompensasi bagi warga yang datanya bocor. Di Eropa, GDPR menjamin hak ini.",
		CreatedAt: now.Add(-18 * time.Hour),
	}
	c6b := commentModels.Comment{
		ID:        uuid.New(),
		UserID:    user2.ID,
		PostID:    post6.ID,
		Content:   "Sebagai seseorang yang paham medis tapi awam IT, saya khawatir data pasien puskesmas juga rentan. Apakah ada audit untuk sistem e-Puskesmas?",
		CreatedAt: now.Add(-16 * time.Hour),
	}

	comments := []*commentModels.Comment{
		&c1, &c1Reply, &c1b,
		&c2, &c2Reply,
		&c3, &c3Reply, &c3b, &c3bReply,
		&c4, &c4Reply,
		&c5,
		&c6, &c6Reply, &c6b,
	}

	for _, c := range comments {
		if err := db.Create(c).Error; err != nil {
			log.Printf("⚠️  Failed to seed comment: %v", err)
		}
	}
	log.Printf("✅ %d sample comments seeded", len(comments))

	// ── 7. Create sample follows ──
	follows := []followModels.Follow{
		{ID: uuid.New(), FollowerID: user2.ID, FollowingID: user1.ID},     // Budi → Rina
		{ID: uuid.New(), FollowerID: user3.ID, FollowingID: user1.ID},     // Siti → Rina
		{ID: uuid.New(), FollowerID: user1.ID, FollowingID: user3.ID},     // Rina → Siti
		{ID: uuid.New(), FollowerID: user3.ID, FollowingID: user2.ID},     // Siti → Budi
		{ID: uuid.New(), FollowerID: moderator.ID, FollowingID: user1.ID}, // Dewi → Rina
		{ID: uuid.New(), FollowerID: moderator.ID, FollowingID: user2.ID}, // Dewi → Budi
		{ID: uuid.New(), FollowerID: user1.ID, FollowingID: moderator.ID}, // Rina → Dewi
	}

	for _, f := range follows {
		if err := db.Create(&f).Error; err != nil {
			log.Printf("⚠️  Failed to seed follow: %v", err)
		}
	}
	log.Printf("✅ %d sample follows seeded", len(follows))

	// ── 8. Create sample bookmarks ──
	bookmarks := []bookmarkModels.Bookmark{
		{ID: uuid.New(), UserID: user1.ID, PostID: post2.ID}, // Rina bookmarks Puskesmas post
		{ID: uuid.New(), UserID: user1.ID, PostID: post3.ID}, // Rina bookmarks Transport post
		{ID: uuid.New(), UserID: user2.ID, PostID: post1.ID}, // Budi bookmarks Kurikulum post
		{ID: uuid.New(), UserID: user2.ID, PostID: post6.ID}, // Budi bookmarks Keamanan Data post
		{ID: uuid.New(), UserID: user3.ID, PostID: post4.ID}, // Siti bookmarks Citarum post
		{ID: uuid.New(), UserID: user3.ID, PostID: post6.ID}, // Siti bookmarks Keamanan Data post
		{ID: uuid.New(), UserID: moderator.ID, PostID: post1.ID},
		{ID: uuid.New(), UserID: moderator.ID, PostID: post3.ID},
	}

	for _, b := range bookmarks {
		if err := db.Create(&b).Error; err != nil {
			log.Printf("⚠️  Failed to seed bookmark: %v", err)
		}
	}
	log.Printf("✅ %d sample bookmarks seeded", len(bookmarks))

	// ── 9. Create sample notifications ──
	notifications := []notifModels.Notification{
		// Rina gets notified: Budi followed her
		{
			ID: uuid.New(), UserID: user1.ID, Type: notifModels.NotifTypeNewFollower,
			ActorID: user2.ID, ReferenceID: user2.ID, ReferenceType: "user",
			Message: "Budi Santoso mulai mengikuti Anda", IsRead: true,
			CreatedAt: now.Add(-100 * time.Hour),
		},
		// Rina gets notified: Siti followed her
		{
			ID: uuid.New(), UserID: user1.ID, Type: notifModels.NotifTypeNewFollower,
			ActorID: user3.ID, ReferenceID: user3.ID, ReferenceType: "user",
			Message: "Siti Aminah mulai mengikuti Anda", IsRead: true,
			CreatedAt: now.Add(-80 * time.Hour),
		},
		// Rina gets notified: Budi commented on her post
		{
			ID: uuid.New(), UserID: user1.ID, Type: notifModels.NotifTypePostComment,
			ActorID: user2.ID, ReferenceID: post1.ID, ReferenceType: "post",
			Message: "Budi Santoso mengomentari post Anda \"Kurikulum Merdeka Perlu Evaluasi...\"", IsRead: false,
			CreatedAt: now.Add(-164 * time.Hour),
		},
		// Budi gets notified: user1 voted his post
		{
			ID: uuid.New(), UserID: user2.ID, Type: notifModels.NotifTypePostVote,
			ActorID: user1.ID, ReferenceID: post2.ID, ReferenceType: "post",
			Message: "Rina Wulandari mendukung post Anda \"Anggaran Puskesmas 2026...\"", IsRead: true,
			CreatedAt: now.Add(-110 * time.Hour),
		},
		// Budi gets notified: Siti commented on his post
		{
			ID: uuid.New(), UserID: user2.ID, Type: notifModels.NotifTypePostComment,
			ActorID: user3.ID, ReferenceID: post2.ID, ReferenceType: "post",
			Message: "Siti Aminah mengomentari post Anda \"Anggaran Puskesmas 2026...\"", IsRead: false,
			CreatedAt: now.Add(-116 * time.Hour),
		},
		// Siti gets notified: Rina followed her
		{
			ID: uuid.New(), UserID: user3.ID, Type: notifModels.NotifTypeNewFollower,
			ActorID: user1.ID, ReferenceID: user1.ID, ReferenceType: "user",
			Message: "Rina Wulandari mulai mengikuti Anda", IsRead: false,
			CreatedAt: now.Add(-60 * time.Hour),
		},
		// Siti gets notified: votes on her post
		{
			ID: uuid.New(), UserID: user3.ID, Type: notifModels.NotifTypePostVote,
			ActorID: user1.ID, ReferenceID: post3.ID, ReferenceType: "post",
			Message: "Rina Wulandari mendukung post Anda \"Integrasi Tiket Transportasi...\"", IsRead: false,
			CreatedAt: now.Add(-90 * time.Hour),
		},
		// Siti gets notified: moderator commented
		{
			ID: uuid.New(), UserID: user3.ID, Type: notifModels.NotifTypePostComment,
			ActorID: moderator.ID, ReferenceID: post3.ID, ReferenceType: "post",
			Message: "Dewi Kartini mengomentari post Anda \"Integrasi Tiket Transportasi...\"", IsRead: false,
			CreatedAt: now.Add(-86 * time.Hour),
		},
	}

	for _, n := range notifications {
		if err := db.Create(&n).Error; err != nil {
			log.Printf("⚠️  Failed to seed notification: %v", err)
		}
	}
	log.Printf("✅ %d sample notifications seeded", len(notifications))

	log.Println("🎉 All sample data seeded successfully!")
}
