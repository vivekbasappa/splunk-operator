// Copyright (c) 2018-2020 Splunk Inc. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// 	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package enterprise

import (
	"errors"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	enterprisev1 "github.com/splunk/splunk-operator/pkg/apis/enterprise/v1alpha2"
	"github.com/splunk/splunk-operator/pkg/splunk/resources"
	"github.com/splunk/splunk-operator/pkg/splunk/spark"
)

// getSplunkVolumeClaims returns a standard collection of Kubernetes volume claims.
func getSplunkVolumeClaims(cr enterprisev1.MetaObject, spec *enterprisev1.CommonSplunkSpec, labels map[string]string) ([]corev1.PersistentVolumeClaim, error) {
	var etcStorage, varStorage resource.Quantity
	var err error

	etcStorage, err = resources.ParseResourceQuantity(spec.EtcStorage, "1Gi")
	if err != nil {
		return []corev1.PersistentVolumeClaim{}, fmt.Errorf("%s: %s", "etcStorage", err)
	}

	varStorage, err = resources.ParseResourceQuantity(spec.VarStorage, "200Gi")
	if err != nil {
		return []corev1.PersistentVolumeClaim{}, fmt.Errorf("%s: %s", "varStorage", err)
	}

	volumeClaims := []corev1.PersistentVolumeClaim{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "etc",
				Namespace: cr.GetNamespace(),
				Labels:    labels,
			},
			Spec: corev1.PersistentVolumeClaimSpec{
				AccessModes: []corev1.PersistentVolumeAccessMode{"ReadWriteOnce"},
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceStorage: etcStorage,
					},
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "var",
				Namespace: cr.GetNamespace(),
				Labels:    labels,
			},
			Spec: corev1.PersistentVolumeClaimSpec{
				AccessModes: []corev1.PersistentVolumeAccessMode{"ReadWriteOnce"},
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceStorage: varStorage,
					},
				},
			},
		},
	}

	if spec.StorageClassName != "" {
		for idx := range volumeClaims {
			volumeClaims[idx].Spec.StorageClassName = &spec.StorageClassName
		}
	}

	return volumeClaims, nil
}

// getSplunkRequirements returns the Kubernetes ResourceRequirements to use for Splunk instances.
func getSplunkRequirements(cr *enterprisev1.SplunkEnterprise) (corev1.ResourceRequirements, error) {
	cpuRequest, err := resources.ParseResourceQuantity(cr.Spec.Resources.SplunkCPURequest, "0.1")
	if err != nil {
		return corev1.ResourceRequirements{}, fmt.Errorf("%s: %s", "SplunkCPURequest", err)
	}

	memoryRequest, err := resources.ParseResourceQuantity(cr.Spec.Resources.SplunkMemoryRequest, "512Mi")
	if err != nil {
		return corev1.ResourceRequirements{}, fmt.Errorf("%s: %s", "SplunkMemoryRequest", err)
	}

	cpuLimit, err := resources.ParseResourceQuantity(cr.Spec.Resources.SplunkCPULimit, "4")
	if err != nil {
		return corev1.ResourceRequirements{}, fmt.Errorf("%s: %s", "SplunkCPULimit", err)
	}

	memoryLimit, err := resources.ParseResourceQuantity(cr.Spec.Resources.SplunkMemoryLimit, "8Gi")
	if err != nil {
		return corev1.ResourceRequirements{}, fmt.Errorf("%s: %s", "SplunkMemoryLimit", err)
	}

	return corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    cpuRequest,
			corev1.ResourceMemory: memoryRequest,
		},
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    cpuLimit,
			corev1.ResourceMemory: memoryLimit,
		}}, nil
}

// getSparkRequirements returns the Kubernetes ResourceRequirements to use for Spark instances.
func getSparkRequirements(cr *enterprisev1.SplunkEnterprise) (corev1.ResourceRequirements, error) {
	cpuRequest, err := resources.ParseResourceQuantity(cr.Spec.Resources.SparkCPURequest, "0.1")
	if err != nil {
		return corev1.ResourceRequirements{}, fmt.Errorf("%s: %s", "SparkCPURequest", err)
	}

	memoryRequest, err := resources.ParseResourceQuantity(cr.Spec.Resources.SparkMemoryRequest, "512Mi")
	if err != nil {
		return corev1.ResourceRequirements{}, fmt.Errorf("%s: %s", "SparkMemoryRequest", err)
	}

	cpuLimit, err := resources.ParseResourceQuantity(cr.Spec.Resources.SparkCPULimit, "4")
	if err != nil {
		return corev1.ResourceRequirements{}, fmt.Errorf("%s: %s", "SparkCPULimit", err)
	}

	memoryLimit, err := resources.ParseResourceQuantity(cr.Spec.Resources.SparkMemoryLimit, "8Gi")
	if err != nil {
		return corev1.ResourceRequirements{}, fmt.Errorf("%s: %s", "SparkMemoryLimit", err)
	}

	return corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    cpuRequest,
			corev1.ResourceMemory: memoryRequest,
		},
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    cpuLimit,
			corev1.ResourceMemory: memoryLimit,
		}}, nil
}

// copyCommonSpec copies common parameters from a SplunkEnterpriseSpec
func copyCommonSpec(dst *enterprisev1.CommonSpec, cr *enterprisev1.SplunkEnterprise, isSpark bool) error {
	dst.Image = cr.Spec.SplunkImage
	dst.ImagePullPolicy = cr.Spec.ImagePullPolicy
	dst.StorageClassName = cr.Spec.StorageClassName
	dst.SchedulerName = cr.Spec.SchedulerName
	dst.Affinity = *cr.Spec.Affinity.DeepCopy()

	var err error
	if isSpark {
		dst.Resources, err = getSparkRequirements(cr)
	} else {
		dst.Resources, err = getSplunkRequirements(cr)
	}
	return err
}

// copyCommonSplunkSpec copies common Splunk Enterprise parameters from a SplunkEnterpriseSpec
func copyCommonSplunkSpec(dst *enterprisev1.CommonSplunkSpec, cr *enterprisev1.SplunkEnterprise, instanceType InstanceType) error {
	dst.EtcStorage = cr.Spec.Resources.SplunkEtcStorage
	if instanceType == SplunkIndexer {
		dst.VarStorage = cr.Spec.Resources.SplunkIndexerStorage
	} else {
		dst.VarStorage = cr.Spec.Resources.SplunkVarStorage
	}
	dst.Volumes = make([]corev1.Volume, len(cr.Spec.SplunkVolumes))
	copy(dst.Volumes, cr.Spec.SplunkVolumes)
	dst.Defaults = cr.Spec.Defaults
	dst.DefaultsURL = cr.Spec.DefaultsURL
	dst.LicenseURL = cr.Spec.LicenseURL
	if dst.LicenseURL != "" {
		dst.LicenseMasterRef = corev1.ObjectReference{
			Name:      cr.GetName(),
			Namespace: cr.GetNamespace(),
		}
	}
	dst.IndexerRef = corev1.ObjectReference{
		Name:      cr.GetName(),
		Namespace: cr.GetNamespace(),
	}
	return copyCommonSpec(&dst.CommonSpec, cr, false)
}

// GetLicenseMasterResource returns corresponding LicenseMaster type managed by a SplunkEnterprise
func GetLicenseMasterResource(cr *enterprisev1.SplunkEnterprise) (*enterprisev1.LicenseMaster, error) {
	result := enterprisev1.LicenseMaster{
		ObjectMeta: metav1.ObjectMeta{
			Name:        cr.GetName(),
			Namespace:   cr.GetNamespace(),
			Labels:      cr.GetObjectMeta().GetLabels(),
			Annotations: cr.GetObjectMeta().GetAnnotations(),
			Finalizers:  cr.GetObjectMeta().GetFinalizers(),
		},
	}
	err := copyCommonSplunkSpec(&result.Spec.CommonSplunkSpec, cr, SplunkLicenseMaster)
	result.SetOwnerReferences(append(result.GetOwnerReferences(), resources.AsOwner(cr)))
	return &result, err
}

// GetStandaloneResource returns corresponding Standalone type managed by a SplunkEnterprise
func GetStandaloneResource(cr *enterprisev1.SplunkEnterprise) (*enterprisev1.Standalone, error) {
	result := enterprisev1.Standalone{
		ObjectMeta: metav1.ObjectMeta{
			Name:        cr.GetName(),
			Namespace:   cr.GetNamespace(),
			Labels:      cr.GetObjectMeta().GetLabels(),
			Annotations: cr.GetObjectMeta().GetAnnotations(),
			Finalizers:  cr.GetObjectMeta().GetFinalizers(),
		},
	}
	result.Spec.Replicas = cr.Spec.Topology.Standalones
	result.Spec.SparkImage = cr.Spec.SparkImage
	if cr.Spec.EnableDFS {
		result.Spec.SparkRef = corev1.ObjectReference{
			Name:      cr.GetName(),
			Namespace: cr.GetNamespace(),
		}
	}
	err := copyCommonSplunkSpec(&result.Spec.CommonSplunkSpec, cr, SplunkStandalone)
	result.SetOwnerReferences(append(result.GetOwnerReferences(), resources.AsOwner(cr)))
	return &result, err
}

// GetSearchHeadResource returns corresponding SearchHead type managed by a SplunkEnterprise
func GetSearchHeadResource(cr *enterprisev1.SplunkEnterprise) (*enterprisev1.SearchHead, error) {
	result := enterprisev1.SearchHead{
		ObjectMeta: metav1.ObjectMeta{
			Name:        cr.GetName(),
			Namespace:   cr.GetNamespace(),
			Labels:      cr.GetObjectMeta().GetLabels(),
			Annotations: cr.GetObjectMeta().GetAnnotations(),
			Finalizers:  cr.GetObjectMeta().GetFinalizers(),
		},
	}
	result.Spec.Replicas = cr.Spec.Topology.SearchHeads
	result.Spec.SparkImage = cr.Spec.SparkImage
	if cr.Spec.EnableDFS {
		result.Spec.SparkRef = corev1.ObjectReference{
			Name:      cr.GetName(),
			Namespace: cr.GetNamespace(),
		}
	}
	err := copyCommonSplunkSpec(&result.Spec.CommonSplunkSpec, cr, SplunkSearchHead)
	result.SetOwnerReferences(append(result.GetOwnerReferences(), resources.AsOwner(cr)))
	return &result, err
}

// GetIndexerResource returns corresponding Indexer type managed by a SplunkEnterprise
func GetIndexerResource(cr *enterprisev1.SplunkEnterprise) (*enterprisev1.Indexer, error) {
	result := enterprisev1.Indexer{
		ObjectMeta: metav1.ObjectMeta{
			Name:        cr.GetName(),
			Namespace:   cr.GetNamespace(),
			Labels:      cr.GetObjectMeta().GetLabels(),
			Annotations: cr.GetObjectMeta().GetAnnotations(),
			Finalizers:  cr.GetObjectMeta().GetFinalizers(),
		},
	}
	result.Spec.Replicas = cr.Spec.Topology.Indexers
	err := copyCommonSplunkSpec(&result.Spec.CommonSplunkSpec, cr, SplunkIndexer)
	result.SetOwnerReferences(append(result.GetOwnerReferences(), resources.AsOwner(cr)))
	return &result, err
}

// GetSparkResource returns corresponding Spark type managed by a SplunkEnterprise
func GetSparkResource(cr *enterprisev1.SplunkEnterprise) (*enterprisev1.Spark, error) {
	result := enterprisev1.Spark{
		ObjectMeta: metav1.ObjectMeta{
			Name:        cr.GetName(),
			Namespace:   cr.GetNamespace(),
			Labels:      cr.GetObjectMeta().GetLabels(),
			Annotations: cr.GetObjectMeta().GetAnnotations(),
			Finalizers:  cr.GetObjectMeta().GetFinalizers(),
		},
	}
	result.Spec.Replicas = cr.Spec.Topology.SparkWorkers
	err := copyCommonSpec(&result.Spec.CommonSpec, cr, true)
	result.SetOwnerReferences(append(result.GetOwnerReferences(), resources.AsOwner(cr)))
	return &result, err
}

// GetStandaloneStatefulSet returns a Kubernetes StatefulSet object for Splunk Enterprise standalone instances.
func GetStandaloneStatefulSet(cr *enterprisev1.Standalone) (*appsv1.StatefulSet, error) {

	// get generic statefulset for Splunk Enterprise objects
	ss, err := getSplunkStatefulSet(cr, &cr.Spec.CommonSplunkSpec, SplunkStandalone, cr.Spec.Replicas, []corev1.EnvVar{})
	if err != nil {
		return nil, err
	}

	// add spark and java mounts to search head containers
	if cr.Spec.SparkRef.Name != "" {
		err := addDFCToPodTemplate(&ss.Spec.Template, cr.Spec.SparkRef, cr.Spec.SparkImage, cr.Spec.ImagePullPolicy, cr.Spec.Replicas > 1)
		if err != nil {
			return nil, err
		}
	}

	return ss, nil
}

// GetSearchHeadStatefulSet returns a Kubernetes StatefulSet object for Splunk Enterprise search heads.
func GetSearchHeadStatefulSet(cr *enterprisev1.SearchHead) (*appsv1.StatefulSet, error) {

	// get search head env variables with deployer
	env := getSearchHeadExtraEnv(cr, cr.Spec.Replicas)
	env = append(env, corev1.EnvVar{
		Name:  "SPLUNK_DEPLOYER_URL",
		Value: GetSplunkServiceName(SplunkDeployer, cr.GetIdentifier(), false),
	})

	// get generic statefulset for Splunk Enterprise objects
	ss, err := getSplunkStatefulSet(cr, &cr.Spec.CommonSplunkSpec, SplunkSearchHead, cr.Spec.Replicas, env)
	if err != nil {
		return nil, err
	}

	// add spark and java mounts to search head containers
	if cr.Spec.SparkRef.Name != "" {
		err := addDFCToPodTemplate(&ss.Spec.Template, cr.Spec.SparkRef, cr.Spec.SparkImage, cr.Spec.ImagePullPolicy, cr.Spec.Replicas > 1)
		if err != nil {
			return nil, err
		}
	}

	return ss, nil
}

// GetIndexerStatefulSet returns a Kubernetes StatefulSet object for Splunk Enterprise indexers.
func GetIndexerStatefulSet(cr *enterprisev1.Indexer) (*appsv1.StatefulSet, error) {
	return getSplunkStatefulSet(cr, &cr.Spec.CommonSplunkSpec, SplunkIndexer, cr.Spec.Replicas, getIndexerExtraEnv(cr, cr.Spec.Replicas))
}

// GetClusterMasterStatefulSet returns a Kubernetes StatefulSet object for a Splunk Enterprise license master.
func GetClusterMasterStatefulSet(cr *enterprisev1.Indexer) (*appsv1.StatefulSet, error) {
	return getSplunkStatefulSet(cr, &cr.Spec.CommonSplunkSpec, SplunkClusterMaster, 1, getIndexerExtraEnv(cr, cr.Spec.Replicas))
}

// GetDeployerStatefulSet returns a Kubernetes StatefulSet object for a Splunk Enterprise license master.
func GetDeployerStatefulSet(cr *enterprisev1.SearchHead) (*appsv1.StatefulSet, error) {
	return getSplunkStatefulSet(cr, &cr.Spec.CommonSplunkSpec, SplunkDeployer, 1, getSearchHeadExtraEnv(cr, cr.Spec.Replicas))
}

// GetLicenseMasterStatefulSet returns a Kubernetes StatefulSet object for a Splunk Enterprise license master.
func GetLicenseMasterStatefulSet(cr *enterprisev1.LicenseMaster) (*appsv1.StatefulSet, error) {
	return getSplunkStatefulSet(cr, &cr.Spec.CommonSplunkSpec, SplunkLicenseMaster, 1, []corev1.EnvVar{})
}

// GetSplunkService returns a Kubernetes Service object for Splunk instances configured for a SplunkEnterprise resource.
func GetSplunkService(cr enterprisev1.MetaObject, spec enterprisev1.CommonSpec, instanceType InstanceType, isHeadless bool) *corev1.Service {

	// use template if not headless
	var service *corev1.Service
	if isHeadless {
		service = &corev1.Service{}
		service.Spec.ClusterIP = corev1.ClusterIPNone
	} else {
		service = spec.ServiceTemplate.DeepCopy()
	}
	service.TypeMeta = metav1.TypeMeta{
		Kind:       "Service",
		APIVersion: "v1",
	}
	service.ObjectMeta.Name = GetSplunkServiceName(instanceType, cr.GetIdentifier(), isHeadless)
	service.ObjectMeta.Namespace = cr.GetNamespace()
	service.Spec.Selector = resources.GetLabels(cr.GetIdentifier(), instanceType.ToString(), true)
	service.Spec.Ports = resources.SortServicePorts(getSplunkServicePorts(instanceType)) // note that port order is important for tests

	// append standard labels
	if service.ObjectMeta.Labels == nil {
		service.ObjectMeta.Labels = make(map[string]string)
	}
	for k, v := range resources.GetLabels(cr.GetIdentifier(), fmt.Sprintf("%s-%s", instanceType, "service"), false) {
		service.ObjectMeta.Labels[k] = v
	}
	// append labels from parent
	for k, v := range cr.GetObjectMeta().GetLabels() {
		service.ObjectMeta.Labels[k] = v
	}

	// append annotations from parent
	if service.ObjectMeta.Annotations == nil {
		service.ObjectMeta.Annotations = make(map[string]string)
	}
	for k, v := range cr.GetObjectMeta().GetAnnotations() {
		service.ObjectMeta.Annotations[k] = v
	}

	if instanceType == SplunkDeployer || (instanceType == SplunkSearchHead && isHeadless) {
		// required for SHC bootstrap process; use services with heads when readiness is desired
		service.Spec.PublishNotReadyAddresses = true
	}

	service.SetOwnerReferences(append(service.GetOwnerReferences(), resources.AsOwner(cr)))

	return service
}

// validateCommonSplunkSpec checks validity and makes default updates to a CommonSplunkSpec, and returns error if something is wrong.
func validateCommonSplunkSpec(spec *enterprisev1.CommonSplunkSpec) error {
	// if not specified via spec or env, image defaults to splunk/splunk
	spec.CommonSpec.Image = GetSplunkImage(spec.CommonSpec.Image)
	defaultResources := corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("0.1"),
			corev1.ResourceMemory: resource.MustParse("512Mi"),
		},
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("4"),
			corev1.ResourceMemory: resource.MustParse("8Gi"),
		},
	}
	// work-around openapi validation error by ensuring it is not nill
	if spec.Volumes == nil {
		spec.Volumes = []corev1.Volume{}
	}
	return resources.ValidateCommonSpec(&spec.CommonSpec, defaultResources)
}

// ValidateIndexerSpec checks validity and makes default updates to a IndexerSpec, and returns error if something is wrong.
func ValidateIndexerSpec(spec *enterprisev1.IndexerSpec) error {
	if spec.Replicas == 0 {
		spec.Replicas = 1
	}
	return validateCommonSplunkSpec(&spec.CommonSplunkSpec)
}

// ValidateSearchHeadSpec checks validity and makes default updates to a SearchHeadSpec, and returns error if something is wrong.
func ValidateSearchHeadSpec(spec *enterprisev1.SearchHeadSpec) error {
	if spec.Replicas < 3 {
		spec.Replicas = 3
	}
	spec.SparkImage = spark.GetSparkImage(spec.SparkImage)
	return validateCommonSplunkSpec(&spec.CommonSplunkSpec)
}

// ValidateStandaloneSpec checks validity and makes default updates to a StandaloneSpec, and returns error if something is wrong.
func ValidateStandaloneSpec(spec *enterprisev1.StandaloneSpec) error {
	if spec.Replicas == 0 {
		spec.Replicas = 1
	}
	spec.SparkImage = spark.GetSparkImage(spec.SparkImage)
	return validateCommonSplunkSpec(&spec.CommonSplunkSpec)
}

// ValidateLicenseMasterSpec checks validity and makes default updates to a LicenseMasterSpec, and returns error if something is wrong.
func ValidateLicenseMasterSpec(spec *enterprisev1.LicenseMasterSpec) error {
	return validateCommonSplunkSpec(&spec.CommonSplunkSpec)
}

// ValidateSplunkEnterpriseSpec checks validity and makes default updates to a SplunkEnterpriseSpec, and returns error if something is wrong.
func ValidateSplunkEnterpriseSpec(spec *enterprisev1.SplunkEnterpriseSpec) error {
	// cluster sanity checks
	if spec.Topology.SearchHeads > 0 && spec.Topology.Indexers <= 0 {
		return errors.New("You must specify how many indexers the cluster should have")
	}
	if spec.Topology.SearchHeads <= 0 && spec.Topology.Indexers > 0 {
		return errors.New("You must specify how many search heads the cluster should have")
	}

	// default to using a single standalone instance
	if spec.Topology.SearchHeads <= 0 && spec.Topology.Indexers <= 0 {
		if spec.Topology.Standalones <= 0 {
			spec.Topology.Standalones = 1
		}
	}

	// default to a single spark worker
	if spec.EnableDFS && spec.Topology.SparkWorkers <= 0 {
		spec.Topology.SparkWorkers = 1
	}

	// make sure SchedulerName is not empty
	if spec.SchedulerName == "" {
		spec.SchedulerName = "default-scheduler"
	}

	// work-around openapi validation error by ensuring it is not nill
	if spec.SplunkVolumes == nil {
		spec.SplunkVolumes = []corev1.Volume{}
	}

	// update spark image
	spec.SparkImage = spark.GetSparkImage(spec.SparkImage)

	return resources.ValidateImagePullPolicy(&spec.ImagePullPolicy)
}

// GetSplunkDefaults returns a Kubernetes ConfigMap containing defaults for a SplunkEnterprise resource.
func GetSplunkDefaults(identifier, namespace string, instanceType InstanceType, defaults string) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      GetSplunkDefaultsName(identifier, instanceType),
			Namespace: namespace,
		},
		Data: map[string]string{
			"default.yml": defaults,
		},
	}
}

// GetSplunkSecrets returns a Kubernetes Secret containing randomly generated default secrets to use for a SplunkEnterprise resource.
func GetSplunkSecrets(cr enterprisev1.MetaObject, instanceType InstanceType, idxcSecret []byte) *corev1.Secret {
	// idxc_secret is option, and may be used to override random generation
	if len(idxcSecret) == 0 {
		idxcSecret = generateSplunkSecret()
	}

	// generate some default secret values to share across the cluster
	secretData := map[string][]byte{
		"hec_token":    generateHECToken(),
		"password":     generateSplunkSecret(),
		"pass4SymmKey": generateSplunkSecret(),
		"idxc_secret":  idxcSecret,
		"shc_secret":   generateSplunkSecret(),
	}
	secretData["default.yml"] = []byte(fmt.Sprintf(`
splunk:
    hec_disabled: 0
    hec_enableSSL: 0
    hec_token: "%s"
    password: "%s"
    pass4SymmKey: "%s"
    idxc:
        secret: "%s"
    shc:
        secret: "%s"
`,
		secretData["hec_token"],
		secretData["password"],
		secretData["pass4SymmKey"],
		secretData["idxc_secret"],
		secretData["shc_secret"]))

	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      GetSplunkSecretsName(cr.GetIdentifier(), instanceType),
			Namespace: cr.GetNamespace(),
		},
		Data: secretData,
	}
}

// generateSplunkSecret returns a randomly generated Splunk secret.
func generateSplunkSecret() []byte {
	return resources.GenerateSecret(secretBytes, 24)
}

// generateHECToken returns a randomly generated HEC token formatted like a UUID.
// Note that it is not strictly a UUID, but rather just looks like one.
func generateHECToken() []byte {
	hecToken := resources.GenerateSecret(hexBytes, 36)
	hecToken[8] = '-'
	hecToken[13] = '-'
	hecToken[18] = '-'
	hecToken[23] = '-'
	return hecToken
}

// getSplunkPorts returns a map of ports to use for Splunk instances.
func getSplunkPorts(instanceType InstanceType) map[string]int {
	result := map[string]int{
		"splunkweb": 8000,
		"splunkd":   8089,
	}

	switch instanceType {
	case SplunkStandalone:
		result["dfccontrol"] = 17000
		result["datarecieve"] = 19000
		result["dfsmaster"] = 9000
		result["hec"] = 8088
		result["s2s"] = 9997
	case SplunkSearchHead:
		result["dfccontrol"] = 17000
		result["datarecieve"] = 19000
		result["dfsmaster"] = 9000
	case SplunkIndexer:
		result["hec"] = 8088
		result["s2s"] = 9997
	}

	return result
}

// getSplunkContainerPorts returns a list of Kubernetes ContainerPort objects for Splunk instances.
func getSplunkContainerPorts(instanceType InstanceType) []corev1.ContainerPort {
	l := []corev1.ContainerPort{}
	for key, value := range getSplunkPorts(instanceType) {
		l = append(l, corev1.ContainerPort{
			Name:          key,
			ContainerPort: int32(value),
			Protocol:      "TCP",
		})
	}
	return l
}

// getSplunkServicePorts returns a list of Kubernetes ServicePort objects for Splunk instances.
func getSplunkServicePorts(instanceType InstanceType) []corev1.ServicePort {
	l := []corev1.ServicePort{}
	for key, value := range getSplunkPorts(instanceType) {
		l = append(l, corev1.ServicePort{
			Name:     key,
			Port:     int32(value),
			Protocol: "TCP",
		})
	}
	return l
}

// getSplunkVolumeMounts returns a standard collection of Kubernetes volume mounts.
func getSplunkVolumeMounts() []corev1.VolumeMount {
	return []corev1.VolumeMount{
		{
			Name:      "pvc-etc",
			MountPath: "/opt/splunk/etc",
		},
		{
			Name:      "pvc-var",
			MountPath: "/opt/splunk/var",
		},
	}
}

// addSplunkVolumeToTemplate modifies the podTemplateSpec object to incorporates an additional VolumeSource.
func addSplunkVolumeToTemplate(podTemplateSpec *corev1.PodTemplateSpec, name string, volumeSource corev1.VolumeSource) {
	podTemplateSpec.Spec.Volumes = append(podTemplateSpec.Spec.Volumes, corev1.Volume{
		Name:         "mnt-splunk-" + name,
		VolumeSource: volumeSource,
	})

	for idx := range podTemplateSpec.Spec.Containers {
		containerSpec := &podTemplateSpec.Spec.Containers[idx]
		containerSpec.VolumeMounts = append(containerSpec.VolumeMounts, corev1.VolumeMount{
			Name:      "mnt-splunk-" + name,
			MountPath: "/mnt/splunk-" + name,
		})
	}
}

// addDFCToPodTemplate modifies the podTemplateSpec object to incorporate support for DFS.
func addDFCToPodTemplate(podTemplateSpec *corev1.PodTemplateSpec, sparkRef corev1.ObjectReference, sparkImage string, imagePullPolicy string, slotsEnabled bool) error {
	// create an init container in the pod, which is just used to populate the jdk and spark mount directories
	containerSpec := corev1.Container{
		Image:           sparkImage,
		ImagePullPolicy: corev1.PullPolicy(imagePullPolicy),
		Name:            "init",
		Command:         []string{"bash", "-c", "cp -r /opt/jdk /mnt && cp -r /opt/spark /mnt"},
		VolumeMounts: []corev1.VolumeMount{
			{Name: "mnt-splunk-jdk", MountPath: "/mnt/jdk"},
			{Name: "mnt-splunk-spark", MountPath: "/mnt/spark"},
		},
		Resources: corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("0.25"),
				corev1.ResourceMemory: resource.MustParse("128Mi"),
			},
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("1"),
				corev1.ResourceMemory: resource.MustParse("512Mi"),
			},
		},
	}
	podTemplateSpec.Spec.InitContainers = append(podTemplateSpec.Spec.InitContainers, containerSpec)

	// add empty jdk and spark mount directories to all of the splunk containers
	emptyVolumeSource := corev1.VolumeSource{
		EmptyDir: &corev1.EmptyDirVolumeSource{},
	}
	addSplunkVolumeToTemplate(podTemplateSpec, "jdk", emptyVolumeSource)
	addSplunkVolumeToTemplate(podTemplateSpec, "spark", emptyVolumeSource)

	// prepare spark master host URL
	sparkMasterHost := spark.GetSparkServiceName(spark.SparkMaster, sparkRef.Name, false)
	if sparkRef.Namespace != "" {
		sparkMasterHost = resources.GetServiceFQDN(sparkRef.Namespace, sparkMasterHost)
	}

	// append DFS env variables to splunk enterprise containers
	dfsEnvVar := []corev1.EnvVar{
		{Name: "SPLUNK_ENABLE_DFS", Value: "true"},
		{Name: "SPARK_MASTER_HOST", Value: sparkMasterHost},
		{Name: "SPARK_MASTER_WEBUI_PORT", Value: "8009"},
		{Name: "SPARK_HOME", Value: "/mnt/splunk-spark"},
		{Name: "JAVA_HOME", Value: "/mnt/splunk-jdk"},
		{Name: "SPLUNK_DFW_NUM_SLOTS_ENABLED", Value: "true"},
	}
	if !slotsEnabled {
		dfsEnvVar[5].Value = "false"
	}
	for idx := range podTemplateSpec.Spec.Containers {
		podTemplateSpec.Spec.Containers[idx].Env = append(podTemplateSpec.Spec.Containers[idx].Env, dfsEnvVar...)
	}

	return nil
}

// getSplunkStatefulSet returns a Kubernetes StatefulSet object for Splunk instances configured for a SplunkEnterprise resource.
func getSplunkStatefulSet(cr enterprisev1.MetaObject, spec *enterprisev1.CommonSplunkSpec, instanceType InstanceType, replicas int, extraEnv []corev1.EnvVar) (*appsv1.StatefulSet, error) {

	// prepare labels and other values
	replicas32 := int32(replicas)
	ports := resources.SortContainerPorts(getSplunkContainerPorts(instanceType)) // note that port order is important for tests
	affinity := resources.AppendPodAntiAffinity(&spec.Affinity, cr.GetIdentifier(), instanceType.ToString())
	labels := resources.GetLabels(cr.GetIdentifier(), instanceType.ToString(), false)
	labels["kind"] = instanceType.ToKind() // add kind to labels
	for k, v := range cr.GetObjectMeta().GetLabels() {
		labels[k] = v
	}
	annotations := resources.GetIstioAnnotations(ports)
	for k, v := range cr.GetObjectMeta().GetAnnotations() {
		annotations[k] = v
	}

	// prepare volume claims
	volumeClaims, err := getSplunkVolumeClaims(cr, spec, labels)
	if err != nil {
		return nil, err
	}
	for idx := range volumeClaims {
		volumeClaims[idx].ObjectMeta.Name = fmt.Sprintf("pvc-%s", volumeClaims[idx].ObjectMeta.Name)
	}

	// create statefulset configuration
	statefulSet := &appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "StatefulSet",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      GetSplunkStatefulsetName(instanceType, cr.GetIdentifier()),
			Namespace: cr.GetNamespace(),
		},
		Spec: appsv1.StatefulSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: resources.GetLabels(cr.GetIdentifier(), instanceType.ToString(), true),
			},
			ServiceName:         GetSplunkServiceName(instanceType, cr.GetIdentifier(), true),
			Replicas:            &replicas32,
			PodManagementPolicy: "Parallel",
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      labels,
					Annotations: annotations,
				},
				Spec: corev1.PodSpec{
					Affinity:      affinity,
					SchedulerName: spec.SchedulerName,
					Containers: []corev1.Container{
						{
							Image:           spec.Image,
							ImagePullPolicy: corev1.PullPolicy(spec.ImagePullPolicy),
							Name:            "splunk",
							Ports:           ports,
							VolumeMounts:    getSplunkVolumeMounts(),
						},
					},
				},
			},
			VolumeClaimTemplates: volumeClaims,
		},
	}

	// update statefulset's pod template with common splunk pod config
	updateSplunkPodTemplateWithConfig(&statefulSet.Spec.Template, cr, spec, instanceType, extraEnv)

	// make SplunkEnterprise object the owner
	statefulSet.SetOwnerReferences(append(statefulSet.GetOwnerReferences(), resources.AsOwner(cr)))

	return statefulSet, nil
}

// updateSplunkPodTemplateWithConfig modifies the podTemplateSpec object based on configuration of the SplunkEnterprise resource.
func updateSplunkPodTemplateWithConfig(podTemplateSpec *corev1.PodTemplateSpec, cr enterprisev1.MetaObject, spec *enterprisev1.CommonSplunkSpec, instanceType InstanceType, extraEnv []corev1.EnvVar) {

	// Add custom volumes to splunk containers
	if spec.Volumes != nil {
		podTemplateSpec.Spec.Volumes = append(podTemplateSpec.Spec.Volumes, spec.Volumes...)
		for idx := range podTemplateSpec.Spec.Containers {
			for v := range spec.Volumes {
				podTemplateSpec.Spec.Containers[idx].VolumeMounts = append(podTemplateSpec.Spec.Containers[idx].VolumeMounts, corev1.VolumeMount{
					Name:      spec.Volumes[v].Name,
					MountPath: "/mnt/" + spec.Volumes[v].Name,
				})
			}
		}
	}

	// add defaults secrets to all splunk containers
	addSplunkVolumeToTemplate(podTemplateSpec, "secrets", corev1.VolumeSource{
		Secret: &corev1.SecretVolumeSource{
			SecretName: GetSplunkSecretsName(cr.GetIdentifier(), instanceType),
		},
	})

	// add inline defaults to all splunk containers
	if spec.Defaults != "" {
		addSplunkVolumeToTemplate(podTemplateSpec, "defaults", corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: GetSplunkDefaultsName(cr.GetIdentifier(), instanceType),
				},
			},
		})
	}

	// update security context
	runAsUser := int64(41812)
	fsGroup := int64(41812)
	podTemplateSpec.Spec.SecurityContext = &corev1.PodSecurityContext{
		RunAsUser: &runAsUser,
		FSGroup:   &fsGroup,
	}

	// use script provided by enterprise container to check if pod is alive
	livenessProbe := &corev1.Probe{
		Handler: corev1.Handler{
			Exec: &corev1.ExecAction{
				Command: []string{
					"/sbin/checkstate.sh",
				},
			},
		},
		InitialDelaySeconds: 300,
		TimeoutSeconds:      30,
		PeriodSeconds:       30,
	}

	// pod is ready if container artifact file is created with contents of "started".
	// this indicates that all the the ansible plays executed at startup have completed.
	readinessProbe := &corev1.Probe{
		Handler: corev1.Handler{
			Exec: &corev1.ExecAction{
				Command: []string{
					"/bin/grep",
					"started",
					"/opt/container_artifact/splunk-container.state",
				},
			},
		},
		InitialDelaySeconds: 10,
		TimeoutSeconds:      5,
		PeriodSeconds:       5,
	}

	// prepare defaults variable
	splunkDefaults := "/mnt/splunk-secrets/default.yml"
	if spec.DefaultsURL != "" {
		splunkDefaults = fmt.Sprintf("%s,%s", splunkDefaults, spec.DefaultsURL)
	}
	if spec.Defaults != "" {
		splunkDefaults = fmt.Sprintf("%s,%s", splunkDefaults, "/mnt/splunk-defaults/default.yml")
	}

	// prepare container env variables
	env := []corev1.EnvVar{
		{Name: "SPLUNK_HOME", Value: "/opt/splunk"},
		{Name: "SPLUNK_START_ARGS", Value: "--accept-license"},
		{Name: "SPLUNK_DEFAULTS_URL", Value: splunkDefaults},
		{Name: "SPLUNK_HOME_OWNERSHIP_ENFORCEMENT", Value: "false"},
		{Name: "SPLUNK_ROLE", Value: instanceType.ToRole()},
	}

	// update variables for licensing, if configured
	if spec.LicenseURL != "" {
		env = append(env, corev1.EnvVar{
			Name:  "SPLUNK_LICENSE_URI",
			Value: spec.LicenseURL,
		})
	}
	if instanceType != SplunkLicenseMaster {
		licenseMasterURL := spec.LicenseMasterURL
		if licenseMasterURL == "" && spec.LicenseMasterRef.Name != "" {
			licenseMasterURL = GetSplunkServiceName(SplunkLicenseMaster, spec.LicenseMasterRef.Name, false)
			if spec.LicenseMasterRef.Namespace != "" {
				licenseMasterURL = resources.GetServiceFQDN(spec.LicenseMasterRef.Namespace, licenseMasterURL)
			}
		}
		if licenseMasterURL != "" {
			env = append(env, corev1.EnvVar{
				Name:  "SPLUNK_LICENSE_MASTER_URL",
				Value: licenseMasterURL,
			})
		}
	}

	// append URL for cluster master, if configured
	var clusterMasterURL string
	if instanceType == SplunkIndexer {
		clusterMasterURL = GetSplunkServiceName(SplunkClusterMaster, cr.GetIdentifier(), false)
	} else if instanceType != SplunkClusterMaster {
		clusterMasterURL = spec.ClusterMasterURL
		if clusterMasterURL == "" && spec.IndexerRef.Name != "" {
			clusterMasterURL = GetSplunkServiceName(SplunkClusterMaster, spec.IndexerRef.Name, false)
			if spec.IndexerRef.Namespace != "" {
				clusterMasterURL = resources.GetServiceFQDN(spec.IndexerRef.Namespace, clusterMasterURL)
			}
		}
	}
	if clusterMasterURL != "" {
		extraEnv = append(extraEnv, corev1.EnvVar{
			Name:  "SPLUNK_CLUSTER_MASTER_URL",
			Value: clusterMasterURL,
		})
	}

	// append any extra variables
	env = append(env, extraEnv...)

	// update each container in pod
	for idx := range podTemplateSpec.Spec.Containers {
		podTemplateSpec.Spec.Containers[idx].Resources = spec.Resources
		podTemplateSpec.Spec.Containers[idx].LivenessProbe = livenessProbe
		podTemplateSpec.Spec.Containers[idx].ReadinessProbe = readinessProbe
		podTemplateSpec.Spec.Containers[idx].Env = env
	}
}

// getSearchHeadExtraEnv returns extra environment variables used by search head clusters
func getSearchHeadExtraEnv(cr enterprisev1.MetaObject, replicas int) []corev1.EnvVar {
	return []corev1.EnvVar{
		{
			Name:  "SPLUNK_SEARCH_HEAD_URL",
			Value: GetSplunkStatefulsetUrls(cr.GetNamespace(), SplunkSearchHead, cr.GetIdentifier(), replicas, false),
		}, {
			Name:  "SPLUNK_SEARCH_HEAD_CAPTAIN_URL",
			Value: GetSplunkStatefulsetURL(cr.GetNamespace(), SplunkSearchHead, cr.GetIdentifier(), 0, false),
		},
	}
}

// getIndexerExtraEnv returns extra environment variables used by search head clusters
func getIndexerExtraEnv(cr enterprisev1.MetaObject, replicas int) []corev1.EnvVar {
	return []corev1.EnvVar{
		{
			Name:  "SPLUNK_INDEXER_URL",
			Value: GetSplunkStatefulsetUrls(cr.GetNamespace(), SplunkIndexer, cr.GetIdentifier(), replicas, false),
		},
	}
}
