package utils

type Session struct {
	Session_UID string
	Ip_Addr     string
	Port        int
	IsActive    bool `default:"true"`
}

type SessionIdentifier struct {
	OS            string
	Username      string
	OsVersion     string
	Hostname      string
	SessionObject Session
	MACAddress    []string
}
