/*
Copyright 2022 TriggerMesh Inc.

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

package slacksource

import (
	corev1 "k8s.io/api/core/v1"

	"knative.dev/eventing/pkg/reconciler/source"
	"knative.dev/pkg/apis"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"

	commonv1alpha1 "github.com/triggermesh/triggermesh/pkg/apis/common/v1alpha1"
	"github.com/triggermesh/triggermesh/pkg/apis/sources/v1alpha1"
	common "github.com/triggermesh/triggermesh/pkg/reconciler"
	"github.com/triggermesh/triggermesh/pkg/reconciler/resource"
)

const (
	envSlackAppID         = "SLACK_APP_ID"
	envSlackSigningSecret = "SLACK_SIGNING_SECRET"
)

// adapterConfig contains properties used to configure the source's adapter.
// These are automatically populated by envconfig.
type adapterConfig struct {
	// Container image
	Image string `default:"gcr.io/triggermesh/slacksource-adapter"`

	// Configuration accessor for logging/metrics/tracing
	configs source.ConfigAccessor
}

// Verify that Reconciler implements common.AdapterBuilder.
var _ common.AdapterBuilder[*servingv1.Service] = (*Reconciler)(nil)

// BuildAdapter implements common.AdapterBuilder.
func (r *Reconciler) BuildAdapter(src commonv1alpha1.Reconcilable, sinkURI *apis.URL) (*servingv1.Service, error) {
	typedSrc := src.(*v1alpha1.SlackSource)

	return common.NewAdapterKnService(src, sinkURI,
		resource.Image(r.adapterCfg.Image),

		resource.VisibilityPublic,

		resource.EnvVars(makeSlackEnvs(typedSrc)...),
		resource.EnvVars(r.adapterCfg.configs.ToEnvVars()...),
	), nil
}

func makeSlackEnvs(src *v1alpha1.SlackSource) []corev1.EnvVar {
	var slackEnvs []corev1.EnvVar

	if appID := src.Spec.AppID; appID != nil {
		slackEnvs = append(slackEnvs, corev1.EnvVar{
			Name:  envSlackAppID,
			Value: *appID,
		})
	}

	if signingSecret := src.Spec.SigningSecret; signingSecret != nil {
		slackEnvs = common.MaybeAppendValueFromEnvVar(slackEnvs,
			envSlackSigningSecret, *signingSecret,
		)
	}

	return slackEnvs
}
