package rally

type OpenStackUser struct {
	Username          string `json:"username"`
	Password          string `json:"password"`
	UserDomainName    string `json:"user_domain_name"`
	ProjectName       string `json:"project_name"`
	ProjectDomainName string `json:"project_domain_name"`
}

type OpenStackDeployment struct {
	AuthURL      string          `json:"auth_url"`
	RegionName   string          `json:"region_name"`
	EndpointType string          `json:"endpoint_type"`
	Users        []OpenStackUser `json:"users"`
}

type Deployment struct {
	OpenStackDeployment `json:"openstack"`
}
