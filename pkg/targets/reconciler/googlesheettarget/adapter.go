/*
Copyright 2021 TriggerMesh Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package googlesheettarget

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"knative.dev/eventing/pkg/reconciler/source"
	"knative.dev/pkg/kmeta"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"

	"github.com/triggermesh/triggermesh/pkg/apis/targets/v1alpha1"
	libreconciler "github.com/triggermesh/triggermesh/pkg/targets/reconciler"
	"github.com/triggermesh/triggermesh/pkg/targets/reconciler/resources"
)

const targetPrefix = "googlesheettarget"

const (
	envCredentialsJSON = "GOOGLE_CREDENTIALS_JSON"
	envSheetID         = "SHEET_ID"
	envDefaultPrefix   = "DEFAULT_SHEET_PREFIX"
)

// TargetAdapterArgs are the arguments needed to create a Target Adapter.
// Every field is required.
type TargetAdapterArgs struct {
	Image   string
	Configs source.ConfigAccessor
	Target  *v1alpha1.GoogleSheetTarget
}

// MakeTargetAdapterKService generates (but does not insert into K8s) the Target Adapter KService.
func MakeTargetAdapterKService(args *TargetAdapterArgs) *servingv1.Service {
	name := kmeta.ChildName(targetPrefix+"-", args.Target.Name)
	env := makeEnv(args)
	genericLabels := libreconciler.MakeGenericLabels(targetPrefix, args.Target.Name)
	ksvcLabels := libreconciler.PropagateCommonLabels(args.Target, genericLabels)
	podLabels := libreconciler.PropagateCommonLabels(args.Target, genericLabels)

	return &servingv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: args.Target.Namespace,
			Name:      name,
			Labels:    ksvcLabels,
			OwnerReferences: []metav1.OwnerReference{
				*kmeta.NewControllerRef(args.Target),
			},
		},
		Spec: servingv1.ServiceSpec{
			ConfigurationSpec: servingv1.ConfigurationSpec{
				Template: servingv1.RevisionTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: podLabels,
					},
					Spec: servingv1.RevisionSpec{
						PodSpec: corev1.PodSpec{
							Containers: []corev1.Container{{
								Image: args.Image,
								Env:   env,
							}},
						},
					},
				},
			},
		},
	}
}

func makeEnv(args *TargetAdapterArgs) []corev1.EnvVar {
	env := []corev1.EnvVar{{
		Name:  resources.EnvNamespace,
		Value: args.Target.Namespace,
	}, {
		Name:  resources.EnvName,
		Value: args.Target.Name,
	}, {
		Name:  resources.EnvMetricsDomain,
		Value: resources.DefaultMetricsDomain,
	}, {
		Name: envCredentialsJSON,
		ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: args.Target.Spec.GoogleServiceAccount.SecretKeyRef,
		},
	}, {
		Name:  envSheetID,
		Value: args.Target.Spec.ID,
	}, {
		Name:  envDefaultPrefix,
		Value: args.Target.Spec.DefaultPrefix,
	}}

	return append(env, libreconciler.MakeObsEnv(args.Configs)...)
}