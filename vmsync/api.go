package vmsync

import (
	"app/base/api"
	"app/base/utils"
	"encoding/json"
	"net/http"
	"time"
)

var (
	PageSize     = utils.GetIntEnv("PAGE_SIZE", 5000)
	VmaasCvesURL = utils.GetEnv("BASE_URL", "http://localhost:8080/api/v3/cves")
)

type APICveRequest struct {
	CveList          []string `json:"cve_list"`
	Page             int      `json:"page"`
	PageSize         int      `json:"page_size"`
	RhOnly           bool     `json:"rh_only"`
	ErrataAssociated bool     `json:"errata_associated"`
}

type APICveResponse struct {
	CveList          map[string]APICve `json:"cve_list"`
	Page             int               `json:"page"`
	PageSize         int               `json:"page_size"`
	Pages            int               `json:"pages"`
	RhOnly           bool              `json:"rh_only"`
	ErrataAssociated bool              `json:"errata_associated"`
}

type APICve struct {
	RedhatURL         string     `json:"redhat_url"`
	SecondaryURL      string     `json:"secondary_url"`
	Synopsis          string     `json:"synopsis"`
	Impact            string     `json:"impact"`
	PublicDate        *time.Time `json:"public_date"`
	ModifiedDate      *time.Time `json:"modified_date"`
	CweList           []string   `json:"cwe_list"`
	Cvss3Score        string     `json:"cvss3_score"`
	Cvss3Metrics      string     `json:"cvss3_metrics"`
	Cvss2Score        string     `json:"cvss2_score"`
	Cvss2Metrics      string     `json:"cvss2_metrics"`
	Description       string     `json:"description"`
	PackageList       []string   `json:"package_list"`
	SourcePackageList []string   `json:"source_package_list"`
	ErrataList        []string   `json:"errata_list"`
}

// getAPICves request CVE list from VMaaS
func getAPICves() (map[string]APICve, error) {
	cveMap := make(map[string]APICve)

	client := &api.Client{HTTPClient: &http.Client{}}
	totalPages := 9999

	// Vmaas indexes pages from 1
	for page := 1; page <= totalPages; page++ {
		vmaasRequest := APICveRequest{
			Page:             page,
			CveList:          []string{".*"},
			PageSize:         PageSize,
			RhOnly:           true,
			ErrataAssociated: true,
		}
		vmaasResponse := APICveResponse{}

		statusCode, err := client.RetryRequest(http.MethodPost, VmaasCvesURL, &vmaasRequest, &vmaasResponse)
		if err != nil {
			logger.Warningf("Request %s %s failed: statusCode=%d, err=%s", http.MethodPost, VmaasCvesURL, statusCode, err)
			return cveMap, err
		}

		cveList, err := json.Marshal(&vmaasResponse.CveList)
		if err != nil {
			logger.Warningf("CVE metadata parsing failed, err=%s", err)
			return cveMap, err
		}

		if err = json.Unmarshal(cveList, &cveMap); err != nil {
			logger.Warningf("CVE metadata parsing failed, err=%s", err)
			return cveMap, err
		}

		totalPages = vmaasResponse.Pages
		logger.Infof("Fetched VMAAS cve list: cves=%d, page=%d/%d", len(cveMap), page, vmaasResponse.Pages)
	}

	return cveMap, nil
}
