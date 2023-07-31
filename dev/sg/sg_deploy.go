package main

import (
	"fmt"
	"os"
	"path"
	"text/template"

	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var deployCommand = &cli.Command{
	Name:        "deploy",
	Usage:       `Generate a Kubernetes manifest for a Sourcegraph deployment`,
	Description: `Internal deployments live in the sourcegraph/infra repository.`,
	UsageText: `
sg deploy --values <path to values file>

Example of a values.yaml file:

name: my-app
image: gcr.io/sourcegraph-dev/my-app:latest
replicas: 1
envvars:
  - name: ricky
    value: foo
  - name: julian
    value: bar
containerPorts:
  - name: frontend
    port: 80
servicePorts:
  - name: http
    port: 80
    targetPort: test # Set to the name or port number of the containerPort you want to expose
dns: dave-app.sgdev.org
`,
	Category: CategoryDev,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "values",
			Usage:    "The path to the values file",
			Required: true,
		},
		&cli.BoolFlag{
			Name:     "dry-run",
			Usage:    "Write the manifest to stdout instead of writing to a file",
			Required: false,
		}},
	Action: func(c *cli.Context) error {
		err := generateConfig(c.String("values"), c.Bool("dry-run"))
		if err != nil {
			return errors.Wrap(err, "generate manifest")
		}
		return nil
	}}

type Values struct {
	Name    string
	Envvars []struct {
		Name  string
		Value string
	}
	Image          string
	Replicas       int
	ContainerPorts []struct {
		Name string
		Port int
	} `yaml:"containerPorts"`
	ServicePorts []struct {
		Name       string
		Port       int
		TargetPort interface{} `yaml:"targetPort"` // This can take a string or int
	} `yaml:"servicePorts"`
	Dns string
}

var k8sTemplate = `# This file was geneated by sg deploy.

apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{.Name}}
spec:
  replicas: {{.Replicas}}
  selector:
    matchLabels:
      app: {{.Name}}
  template:
    metadata:
      labels:
        app: {{.Name}}
    spec:
      containers:
        - name: {{.Name}}
          image: {{.Image}}
          imagePullPolicy: Always
          env:
            {{- range $i, $envvar := .Envvars }}
            - name: {{ $envvar.Name }}
              value: {{ $envvar.Value }}
            {{- end }}
          ports:
            {{- range $i, $port := .ContainerPorts }}
            - containerPort: {{ $port.Port }}
              name: {{ $port.Name }}
            {{- end }}
{{ if .ServicePorts -}}
---
apiVersion: v1
kind: Service
metadata:
  name: {{.Name}}-service
spec:
  selector:
    app: {{.Name}}
  ports:
  {{- range $i, $port := .ServicePorts }}
    - port: {{ $port.Port }}
      name: {{ $port.Name }}
      targetPort: {{ $port.TargetPort }}
      protocol: TCP
  {{- end }}
{{- end}}
{{ if .Dns -}}
---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: {{.Name}}-ingress
  namespace: tooling
  annotations:
    kubernetes.io/ingress.class: 'nginx'
spec:
  tls:
    - hosts:
        - {{.Dns}}
      secretName: sgdev-tls-secret
  rules:
    - host: {{.Dns}}
      http:
        paths:
          - backend:
              service:
                name: {{ .Name }}-service
                port:
                  number: {{ (index .ServicePorts 0).Port }}
            path: /
            pathType: Prefix
{{- end }}
`

var dnsTemplate = `
{{- if .Dns -}}
# This file was generated by sg deploy.

locals {
  dogfood_ingress_ip = "34.132.81.184"  # https://github.com/sourcegraph/infrastructure/pull/2125#issuecomment-689637766
}

resource "cloudflare_record" "{{ .Name }}-sgdev-org" {
  zone_id = data.cloudflare_zones.sgdev_org.zones[0].id
  name    = "{{ .Name }}"
  type    = "A"
  value   = local.dogfood_ingress_ip
  proxied = true
}
{{- end }}
`

func generateConfig(configFile string, dryRun bool) error {

	var values Values
	v, err := os.ReadFile(configFile)
	if err != nil {
		return errors.Wrap(err, "read values file")
	}

	err = yaml.Unmarshal(v, &values)
	if err != nil {
		return errors.Wrapf(err, "error unmarshalling values from %q", valuesFile)
	}

	if dryRun {

		fmt.Printf("This is a dry run. The following files would be created:\n\n")
		t := template.Must(template.New("k8s").Parse(k8sTemplate))
		err = t.Execute(os.Stdout, &values)
		if err != nil {
			return errors.Wrap(err, "execute k8s template")
		}
		t = template.Must(template.New("dns").Parse(dnsTemplate))
		err = t.Execute(os.Stdout, &values)
		if err != nil {
			return errors.Wrap(err, "execute dns template")
		}
		return nil
	}

	err = checkCurrentDir("infrastructure")
	if err != nil {
		return err
	}

	k8sPath := fmt.Sprintf("dogfood/kubernetes/tooling/%s", values.Name)
	err = os.MkdirAll(k8sPath, 0755)
	if err != nil {
		return errors.Wrap(err, "create directory")
	}
	k8sOutput, err := os.Create(fmt.Sprintf("%s/%s.yaml", k8sPath, values.Name))
	if err != nil {
		return errors.Wrap(err, "create file")
	}
	defer k8sOutput.Close()

	dnsPath := fmt.Sprintf("dns/%s.sgdev.tf", values.Name)
	dnsOutput, err := os.Create(dnsPath)
	if err != nil {
		return errors.Wrap(err, "create file")
	}
	defer dnsOutput.Close()
	t := template.Must(template.New("k8s").Parse(k8sTemplate))
	err = t.Execute(k8sOutput, &values)
	if err != nil {
		return errors.Wrap(err, "execute k8s template")
	}
	t = template.Must(template.New("dns").Parse(dnsTemplate))
	err = t.Execute(dnsOutput, &values)
	if err != nil {
		return errors.Wrap(err, "execute dns template")
	}

	return nil
}

func checkCurrentDir(expected string) error {

	cwd, err := os.Getwd()
	if err != nil {
		return errors.Wrap(err, "error getting current directory")
	}

	current := path.Base(cwd)
	if current != expected {
		return errors.New("Incorrect directory. Please run from the sourcegraph/infrastructure repository")
	}
	return nil
}
