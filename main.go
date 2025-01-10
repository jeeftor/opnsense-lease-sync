package main

import (
	"flag"
	"fmt"
	"log"
	"opnsense-lease-sync/internal/logger"
	"opnsense-lease-sync/internal/service"
	"os"
)

func main() {
	adguardURL := flag.String("adguard-url", "http://localhost:3000", "AdGuard Home API URL")
	leasePath := flag.String("lease-path", "/var/dhcpd/var/db/dhcpd.leases", "Path to DHCP lease file")
	flag.Parse()

	logger, err := logger.NewLogger()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	syncService := service.New(*adguardURL, *leasePath, logger)

	if err := syncService.Run(); err != nil {
		logger.Error(fmt.Sprintf("Service error: %v", err))
		os.Exit(1)
	}
}
