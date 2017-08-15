package main

import (
    "flag"
    "os"
    "fmt"
    "strconv"
    "encoding/json"
    "reflect"
    "io/ioutil"
    "gopkg.in/yaml.v2"
)

type flagOptArgs struct {
    Config        *InstanceConfig
    HostPort      *string
    AdminPort     *string
    StatsdHost    *string
    StatsdPort    *string
    MakoServiceId *string
    MakoEnv       *string
    MakoVer       *string
}

// IsZeroOrNil uses reflection to determin whether or not any interface x
// is either a Zero value (not to be confused with int 0) or nil
func IsZeroOrNil(x interface{}) bool {
    if x.(reflect.Value).Kind() == reflect.Ptr {
        if x.(reflect.Value).IsNil() {
            return true
        }
        intrfc := x.(reflect.Value).Interface()
        if value, ok := intrfc.(*string); ok {
            if *value == "" {
                return true
            }
        }
        //return x.(reflect.Value).IsNil() || *x.(reflect.Value).Interface().(*string) == ""
    }
    return x == reflect.DeepEqual(x, reflect.Zero(reflect.TypeOf(x)).Interface())
}

type InstanceConfig struct {
    IdpMetadata    string `yaml:"idp_metadata"`
    IdpIssuer      string `yaml:"idp_issuer"`
    IdpApiAuthUrl  string `yaml:"idp_api_auth_url"`
    SamlCallback   string `yaml:"saml_callback"`
    AudienceURI    string `yaml:"audience_uri"`
    CertAuthority  string `yaml:"cert_authority"`
    CommonName     string `yaml:"common_name"`
    EmailAddress   string `yaml:"email_address"`
    CertOrg        string `yaml:"certificate_org"`
}

func (c *InstanceConfig) Parse(config string) error {
    var data []byte
    data, err := ioutil.ReadFile(config)
    if err != nil {
        Logs.Errorf("%+v",err)
        panic(err)
    }
    if err := yaml.Unmarshal(data, c); err != nil {
        Logs.Errorf("%+v", err)
        return err
    }
    if c.IdpMetadata == "" {
        Logs.Errorf("Lemur-client config: invalid cert")
    }
    if c.IdpIssuer == "" {
        Logs.Errorf("Lemur-client config: invalid idp issuer")
    }
    if c.IdpApiAuthUrl == "" {
        Logs.Errorf("Lemur-client config: invalid auth url")
    }
    if c.SamlCallback == "" {
        Logs.Errorf("Lemur-client config: invalid saml callback")
    }
    if c.AudienceURI == "" {
        Logs.Errorf("Lemur-client config: invalid audience uri")
    }
    if c.CertAuthority == "" {
        Logs.Errorf("Lemur-client config: invalid cert authority")
    }
    if c.CommonName == "" {
        Logs.Errorf("Lemur-client config: invalid common name")
    }
    if c.EmailAddress == "" {
        Logs.Errorf("Lemur-client config: invalid email address")
    }
    return nil
}

// MergeInPlace uses reflection to iterate over flagOptArgs structs and use
// possibly-set values in the argument struct to update values in the callee
// struct
func (to *flagOptArgs) MergeInPlace(from *flagOptArgs) {
    for i := 0; i < reflect.TypeOf(to).Elem().NumField(); i++ {
        if x := reflect.ValueOf(from).Elem().Field(i); !IsZeroOrNil(x) {
            if x.String() != "" {
                reflect.ValueOf(to).Elem().Field(i).Set(x)
            }
        }
    }
}

func GetFlags() (*flagOptArgs) {
    flags := &flagOptArgs{Config:        &InstanceConfig{}, // Must overwrite this when config is found
                          HostPort:      flag.String("host_port", "8080", "webserver port"),
                          AdminPort:     flag.String("admin_port", "8081", "admin webserver port"),
                          StatsdHost:    flag.String("statsd_host", "localhost", "statsd host"),
                          StatsdPort:    flag.String("statsd_port", "8125", "statsd port"),
                          MakoServiceId: flag.String("mako_service_id", "lemur-client", "MAKO Service ID"),
                          MakoEnv:       flag.String("mako_environment", "develop", "MAKO Environment"),
                          MakoVer:       flag.String("mako_version", "", "MAKO Version")}
    configyaml := "config.yaml"
    configPath := flag.String("config", configyaml, "Path to config yaml")
    flag.Parse()
    var config InstanceConfig
    if err := config.Parse(*configPath); err != nil {
        Logs.Errorf("Unable to parse config file!")
        panic(err)
    }
    flags.Config = &config
    // Unfortunately there's no good way around using temp variables and the
    // struct members must be pointers
    hostport := os.Getenv("HOST_PORT")
    adminport := os.Getenv("ADMIN_PORT")
    makostatsdhost := os.Getenv("MAKO_STATSD_HOST")
    makostatsdport := os.Getenv("MAKO_STATSD_PORT")
    makosvcid := os.Getenv("MAKO_SERVICE_ID")
    makoenv := os.Getenv("MAKO_ENVIRONMENT")
    makover := os.Getenv("MAKO_VERSION")
    envFlags := &flagOptArgs{Config:        &config,
                             HostPort:      &hostport,
                             AdminPort:     &adminport,
                             StatsdHost:    &makostatsdhost,
                             StatsdPort:    &makostatsdport,
                             MakoServiceId: &makosvcid,
                             MakoEnv:       &makoenv,
                             MakoVer:       &makover}
    // Environment variables take precedence. Update flags to obtain values
    flags.MergeInPlace(envFlags)
    if flags.HostPort == flags.AdminPort {
        Logs.Infof("Cannot create application and admin services on same port. Setting admin to host+1")
        intport, err := strconv.Atoi(*flags.HostPort)
        if err != nil {
            Logs.Errorf("%+v", err)
            panic(err)
        }
        *flags.AdminPort = string(intport + 1)
    }
    // Expect port to be entered as "8080" and not ":8080"
    *flags.HostPort = fmt.Sprintf(":%s", *flags.HostPort)
    // Expect port to be entered as "8081" and not ":8081"
    *flags.AdminPort = fmt.Sprintf(":%s", *flags.AdminPort)
    return flags
}

func (f *flagOptArgs) String() string {
    str, err := json.Marshal(f)
    if err != nil {
        Logs.Errorf("Unable to JSON Marshal optargs struct!")
    }
    return string(str)
}
