package namecheap

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var (
	errTooManyRetries = fmt.Errorf("exceeded max retry limit")
)

// These are the "Auto" TTL settings in Namecheap
const (
	ncDefaultTTL     int           = 1799
	ncDefaultMXPref  int           = 10
	ncDefaultTimeout time.Duration = 30
)

// Provider returns a terraform.ResourceProvider.
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"username": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("NAMECHEAP_USERNAME", nil),
				Description: "A registered username for namecheap",
			},

			"api_user": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("NAMECHEAP_API_USER", nil),
				Description: "A registered apiuser for namecheap",
			},

			"token": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("NAMECHEAP_TOKEN", nil),
				Description: "The token key for API operations.",
			},

			"ip": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("NAMECHEAP_IP", nil),
				Description: "IP addess of the machine running terraform",
			},

			"use_sandbox": {
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("NAMECHEAP_USE_SANDBOX", false),
				Description: "If true, use the namecheap sandbox",
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"namecheap_record": resourceNameCheapRecord(),
			"namecheap_ns":     resourceNameCheapNS(),
		},

		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	log.Printf("[ROB] NAMECHEAP_USE_SANDBOX: %s; use_sandbox: %v", os.Getenv("NAMECHEAP_USE_SANDBOX"), d.Get("use_sandbox"))

	config := Config{
		username:   d.Get("username").(string),
		apiUser:    d.Get("api_user").(string),
		token:      d.Get("token").(string),
		ip:         d.Get("ip").(string),
		useSandbox: d.Get("use_sandbox").(bool),
	}

	return config.Client()
}

// retryAPICall attempts a specific calllback several times with greater pause between attempts.
// The callback should be responsible for modifying state and cleaning up any resources.
func retryAPICall(f func() error) error {
	attempts, max := 0, 5
	for {
		attempts++
		if attempts > max {
			log.Printf("[ERROR] API Retry Limit Reached.")
			return errTooManyRetries
		}
		if err := f(); err != nil {
			log.Printf("[INFO] Err: %v", err.Error())
			if strings.Contains(err.Error(), "expected element type <ApiResponse> but have <html>") {
				log.Printf("[WARN] Bad Namecheap API response received, backing off for %d seconds...", attempts)
				time.Sleep(time.Duration(attempts) * time.Second)
				continue // retry
			}
			return fmt.Errorf("Failed to create namecheap Record: %s", err)
		}

		return nil
	}
}
