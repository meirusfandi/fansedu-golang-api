package mail

import "fmt"

// Rekening tujuan (placeholder; bisa dipindah ke config)
const (
	BankBCA    = "BCA 1234567890 a.n. FansEdu"
	BankMandiri = "Mandiri 0987654321 a.n. FansEdu"
)

// OrderConfirmationBody returns plain text body untuk email konfirmasi pesanan.
// registerLink empty = user sudah punya akun.
func OrderConfirmationBody(orderID, programName string, totalRupiah int, confirmationCode, registerLink string) string {
	transferAmount := totalRupiah
	if confirmationCode != "" {
		// Kode unik 3 digit (0-999) ditambahkan ke nominal transfer
		var code int
		fmt.Sscanf(confirmationCode, "%d", &code)
		transferAmount = totalRupiah + code
	}
	body := fmt.Sprintf(`Konfirmasi Pesanan

Order ID: %s
Program: %s
Total yang harus dibayar: Rp %d

Rekening tujuan transfer:
- %s
- %s

Kode unik untuk verifikasi: %s
Nominal transfer: Rp %d (total + kode unik)

Cara transfer:
1. Transfer ke salah satu rekening di atas
2. Masukkan nominal persis Rp %d
3. Simpan bukti transfer dan upload di halaman konfirmasi

`,
		orderID, programName, totalRupiah,
		BankBCA, BankMandiri,
		confirmationCode, transferAmount, transferAmount)
	if registerLink != "" {
		body += fmt.Sprintf("Anda belum memiliki akun. Buat akun untuk akses program setelah pembayaran terverifikasi:\n%s\n", registerLink)
	}
	return body
}

// PaymentProofReceivedBody returns body untuk email "Bukti Pembayaran Diterima".
func PaymentProofReceivedBody(orderID, programName string) string {
	return fmt.Sprintf(`Bukti Pembayaran Diterima

Order ID: %s
Program: %s
Status: Menunggu verifikasi

Estimasi waktu verifikasi: 1x24 jam.
Kami akan mengirim email lagi setelah pembayaran terverifikasi.
`,
		orderID, programName)
}

// PaymentVerifiedBody returns body untuk email "Pembayaran Terverifikasi".
func PaymentVerifiedBody(programName, registerLink string) string {
	body := fmt.Sprintf(`Pembayaran Terverifikasi

Anda sudah terdaftar di program: %s
Silakan login untuk mulai belajar.
`, programName)
	if registerLink != "" {
		body += "\nBelum punya akun? Buat akun di: " + registerLink + "\n"
	}
	return body
}
