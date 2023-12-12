package model

import bs "janus-idp.io/backstage-operator/api/v1alpha1"

type DetailedBackstageSpec struct {
	bs.BackstageSpec
	Details SpecDetails
}

type SpecDetails struct {
	AppConfigs             []AppConfigDetails
	ExtraSecretsToFiles    []ExtraSecretToFilesDetails
	ExtraSecretsToEnvs     []ExtraSecretToEnvsDetails
	ExtraConfigMapsToFiles []ExtraConfigMapToFilesDetails
	ExtraConfigMapsToEnvs  []ExtraConfigMapToEnvsDetails
}

type AppConfigDetails struct {
	ConfigMapName string
	FilePath      string
}

type ExtraSecretToFilesDetails struct {
	SecretName string
	FilePaths  []string
}

type ExtraSecretToEnvsDetails struct {
	SecretName string
	Envs       []string
}

type ExtraConfigMapToFilesDetails struct {
	ConfigMapName string
	FilePaths     []string
}

type ExtraConfigMapToEnvsDetails struct {
	ConfigMapName string
	Envs          []string
}
