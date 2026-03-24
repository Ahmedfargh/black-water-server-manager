package Services

import (
	"sync"

	http "net/http"

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
	resp, err := http.Get(site.Health_Route)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		result, err := s.crud_service.AddAnalytics(&Models.SiteHealthStatus{
			SiteID: site.ID,
			Status: "up",
			Time:   time.Now().Format("2006-01-02 15:04:05"),
		}, site.ID)
		if err != nil {
			return
		} else {
			channel <- *result
		}
	} else if resp.StatusCode >= 300 && resp.StatusCode < 400 {
		result, err := s.crud_service.AddAnalytics(&Models.SiteHealthStatus{
			SiteID: site.ID,
			Status: "redirection",
			Time:   time.Now().Format("2006-01-02 15:04:05"),
		}, site.ID)
		if err != nil {
			return
		} else {
			channel <- *result
		}
	} else if resp.StatusCode >= 400 && resp.StatusCode < 500 {
		result, err := s.crud_service.AddAnalytics(&Models.SiteHealthStatus{
			SiteID: site.ID,
			Status: "not_found",
			Time:   time.Now().Format("2006-01-02 15:04:05"),
		}, site.ID)
		if err != nil {
			return
		} else {
			channel <- *result
		}
	} else if resp.StatusCode >= 500 && resp.StatusCode < 600 {
		result, err := s.crud_service.AddAnalytics(&Models.SiteHealthStatus{
			SiteID: site.ID,
			Status: "server_error",
			Time:   time.Now().Format("2006-01-02 15:04:05"),
		}, site.ID)
		if err != nil {
			return
		} else {
			channel <- *result
		}
	}
}
func (s *SiteHealthService) RunCheckUp() (map[uint][]Models.SiteHealthStatus, error) {
	chunk := 5
	page := 1
	results := make(map[uint][]Models.SiteHealthStatus)
	for {
		sites, _, err := s.crud_service.GetSites(page, chunk)
		if err != nil {
			return nil, err
		}
		if len(sites) == 0 {
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
			result := <-chans[i]
			results[site.ID] = append(results[site.ID], result)
			close(chans[i])
		}

		page++
	}
	return results, nil
}
