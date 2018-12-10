/*
Copyright 2018 The Kubernetes Authors.

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

package gracefulshutdown

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/parnurzeal/gorequest"
	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.IngressNginxDescribe("Graceful Shutdown - Slow Requests", func() {
	f := framework.NewDefaultFramework("shutdown-slow-requests")

	BeforeEach(func() {
		f.NewSlowEchoDeployment()
	})

	AfterEach(func() {
	})

	It("should let slow requests finish before shutting down", func() {
		host := "graceful-shutdown"

		f.EnsureIngress(framework.NewSingleIngress(host, "/", host, f.IngressController.Namespace, "http-svc", 80, nil))
		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, fmt.Sprintf("server_name %s ;", host))
			})

		done := make(chan bool)
		go func() {
			defer GinkgoRecover()
			framework.Logf("MAKING SLOW REQUEST")
			resp, _, errs := gorequest.New().
				Get(f.IngressController.HTTPURL).
				Set("Host", host).
				End()
			framework.Logf("   finished slow request")
			Expect(errs).To(BeNil())
			Expect(resp.StatusCode).Should(Equal(http.StatusOK))
			done <- true
		}()

		time.Sleep(250 * time.Millisecond)
		f.DeleteNGINXPod()
		framework.Logf("pod deleted (should happen before finished slow reqs)")
		<-done
	})
})
