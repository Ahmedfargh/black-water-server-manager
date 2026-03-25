package Services

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	config "github.com/ahmedfargh/server-manager/Config"
	CRUD "github.com/ahmedfargh/server-manager/Database/CRUD"
	Models "github.com/ahmedfargh/server-manager/Database/Models"
	Repository "github.com/ahmedfargh/server-manager/Database/Repository"
)

type SiteHealthService struct {
	crud_service *CRUD.SiteCrud
}

func NewSiteHealthService() *SiteHealthService {
	return &SiteHealthService{
		crud_service: CRUD.NewSiteCrud(Repository.NewSiteRepository(config.DB)),
	}
}

func (s *SiteHealthService) CheckSite(site *Models.Site, wg *sync.WaitGroup, channel chan Models.SiteHealthStatus) {
	defer wg.Done()

	client := http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(site.Health_Route)

	status := "down"

	if err == nil {
		defer resp.Body.Close()
		code := resp.StatusCode
		switch {
		case code >= 200 && code < 300:
			status = "up"
		case code >= 300 && code < 400:
			status = "redirection"
		case code >= 400 && code < 500:
			status = "not_found"
		case code >= 500:
			status = "server_error"
		}
	}

	healthData := &Models.SiteHealthStatus{
		SiteID: site.ID,
		Status: status,
		Time:   time.Now().Format("2006-01-02 15:04:05"),
	}

	result, err := s.crud_service.AddAnalytics(healthData, site.ID)
	if err == nil && result != nil {
		channel <- *result
	} else {

	}
}

func (s *SiteHealthService) RunCheckUp() (map[uint][]Models.SiteHealthStatus, error) {
	chunk := 10
	page := 1
	results := make(map[uint][]Models.SiteHealthStatus)

	for {
		sites, _, err := s.crud_service.GetSites(page, chunk)
		if err != nil || len(sites) == 0 {
			break
		}

		chans := make([]chan Models.SiteHealthStatus, len(sites))
		var wg sync.WaitGroup

		for i, site := range sites {
			chans[i] = make(chan Models.SiteHealthStatus, 1)
			wg.Add(1)
			go s.CheckSite(&site, &wg, chans[i])
		}

		wg.Wait()

		for i, site := range sites {
			select {
			case result := <-chans[i]:
				results[site.ID] = append(results[site.ID], result)
			default:
			}
			close(chans[i])
		}
		page++
	}
	return results, nil
}

func (s *SiteHealthService) GetSiteStatusReport(site_id uint, start_date string, end_date string) (map[string]interface{}, error) {
	results, err := s.crud_service.Rep.GetSiteHealthStatusByDate(site_id, start_date, end_date)
	if err != nil {
		return nil, err
	}

	stats := map[string]int{
		"up":           0,
		"redirection":  0,
		"not_found":    0,
		"server_error": 0,
		"down":         0,
	}

	for _, r := range results {
		if _, ok := stats[r.Status]; ok {
			stats[r.Status]++
		} else {
			stats["down"]++
		}
	}

	total := len(results)
	uptimePercent := 0.0
	if total > 0 {
		uptimePercent = (float64(stats["up"]) / float64(total)) * 100
	}

	report := map[string]interface{}{
		"total":              total,
		"total_up":           stats["up"],
		"total_redirection":  stats["redirection"],
		"total_not_found":    stats["not_found"],
		"total_server_error": stats["server_error"],
		"total_down":         stats["down"],
		"uptime_percentage":  fmt.Sprintf("%.2f%%", uptimePercent),
	}
	return report, nil
}
func (s *SiteHealthService) GetSiteHealthStatus(site_id uint, page int, limit int) ([]Models.SiteHealthStatus, uint, error) {
	results, total, err := s.crud_service.GetSiteHealthStatus(site_id, page, limit)

	if err != nil {
		return nil, 0, err
	}

	return results, total, nil
}
