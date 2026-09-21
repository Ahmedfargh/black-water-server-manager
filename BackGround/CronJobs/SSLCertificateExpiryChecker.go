package cronjobs

import (
	"fmt"
	"time"

	Config "github.com/ahmedfargh/server-manager/Config"
	models "github.com/ahmedfargh/server-manager/Database/Models"
	"github.com/ahmedfargh/server-manager/Drivers/NotificationDrivers"
	sslservice "github.com/ahmedfargh/server-manager/Services/SSL"
)

type SSLCertificateExpiryChecker struct {
	CheckInterval time.Duration
}

func (s *SSLCertificateExpiryChecker) Run() (int32, error) {
	interval := s.CheckInterval
	if interval <= 0 {
		interval = 6 * time.Hour
	}

	ticker := time.NewTicker(interval)
	driver := &NotificationDrivers.DiscordDriver{}
	sslSvc := sslservice.NewSSLService()

	for {
		s.checkSites(sslSvc, driver)
		<-ticker.C
	}
}

func (s *SSLCertificateExpiryChecker) checkSites(sslSvc *sslservice.SSLService, driver NotificationDrivers.NotificationInterface) {
	var sites []models.Site
	if err := Config.DB.Find(&sites).Error; err != nil {
		fmt.Printf("[SSLCertChecker] Failed to load sites: %v\n", err)
		return
	}

	for _, site := range sites {
		info, err := sslSvc.CheckSiteSSLAndUpdate(site.ID)
		if err != nil {
			continue
		}

		if info != nil && info.IsExpiringSoon {
			msg := fmt.Sprintf("⚠️ SSL Certificate for %s expires in %d days (%s)", site.Name, info.DaysRemaining, info.NotAfter.Format("2006-01-02"))
			meta := map[string]string{
				"title":       "SSL Expiration Warning",
				"description": msg,
				"bot_token":   Config.GetKey("DISCORD_BOT_TOKEN"),
				"channel_id":  Config.GetKey("DISCORD_CHANNEL_ID"),
			}
			go driver.Send("SSL Monitor", msg, meta)

			// If Auto-Renew is enabled, attempt automatic renewal
			if site.SSLAutoRenew {
				fmt.Printf("[SSLCertChecker] Triggering auto-renewal for site: %s\n", site.Name)
				domain := info.SubjectCommonName
				if domain == "" && len(info.SANs) > 0 {
					domain = info.SANs[0]
				}
				if domain != "" {
					_, _ = sslSvc.RequestCertbotCert(domain, "", "", false)
				}
			}
		}
	}
}

func (s *SSLCertificateExpiryChecker) HandleError(err error) error {
	fmt.Printf("[SSLCertChecker] Error: %v\n", err)
	return err
}
