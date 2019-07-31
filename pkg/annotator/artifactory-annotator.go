/*
Copyright (C) 2019 Synopsys, Inc.

Licensed to the Apache Software Foundation (ASF) under one
or more contributor license agreements. See the NOTICE file
distributed with this work for additional information
regarding copyright ownership. The ASF licenses this file
to you under the Apache License, Version 2.0 (the
"License"); you may not use this file except in compliance
with the License. You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing,
software distributed under the License is distributed on an
"AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
KIND, either express or implied. See the License for the
specific language governing permissions and limitations
under the License.
*/

package annotator

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/blackducksoftware/perceivers/pkg/communicator"
	"github.com/blackducksoftware/perceivers/pkg/metrics"
	"github.com/blackducksoftware/perceivers/pkg/utils"

	perceptorapi "github.com/blackducksoftware/perceptor/pkg/api"
	log "github.com/sirupsen/logrus"
)

// BlackDuck Annotation names
const (
	bdPolicy = "blackduck.policyViolations"
	bdVuln   = "blackduck.vulnerabilities"
	bdSt     = "blackduck.overallStatus"
	bdComp   = "blackduck.componentsURL"
)

// ReposBySha collects URIs for given SHA256
type ReposBySha struct {
	Results []struct {
		URI string `json:"uri"`
	} `json:"results"`
}

// ArtifactoryAnnotator handles annotating artifactory images with vulnerability and policy issues
type ArtifactoryAnnotator struct {
	scanResultsURL string
	registryAuths  []*utils.RegistryAuth
}

// NewArtifactoryAnnotator creates a new ArtifactoryAnnotator object
func NewArtifactoryAnnotator(perceptorURL string, registryAuths []*utils.RegistryAuth) *ArtifactoryAnnotator {
	return &ArtifactoryAnnotator{
		scanResultsURL: fmt.Sprintf("%s/%s", perceptorURL, perceptorapi.ScanResultsPath),
		registryAuths:  registryAuths,
	}
}

// Run starts a controller that will annotate images
func (ia *ArtifactoryAnnotator) Run(interval time.Duration, stopCh <-chan struct{}) {
	log.Infof("starting artifactory annotator controller")

	for {
		select {
		case <-stopCh:
			return
		default:
		}

		time.Sleep(interval)

		err := ia.annotate()
		if err != nil {
			log.Errorf("failed to annotate images: %v", err)
		}
	}
}

func (ia *ArtifactoryAnnotator) annotate() error {
	// Get all the scan results from the Perceptor
	log.Infof("attempting to GET %s for artifactory image annotation", ia.scanResultsURL)
	scanResults, err := ia.getScanResults()
	if err != nil {
		metrics.RecordError("artifactory_annotator", "error getting scan results")
		return fmt.Errorf("error getting scan results: %v", err)
	}

	// Process the scan results and apply annotations/labels to images
	log.Infof("GET to %s succeeded, about to update annotations on all artifactory images", ia.scanResultsURL)
	ia.addAnnotationsToImages(*scanResults)
	return nil
}

func (ia *ArtifactoryAnnotator) getScanResults() (*perceptorapi.ScanResults, error) {
	var results perceptorapi.ScanResults

	bytes, err := communicator.GetPerceptorScanResults(ia.scanResultsURL)
	if err != nil {
		metrics.RecordError("artifactory_annotator", "unable to get scan results")
		return nil, fmt.Errorf("unable to get scan results: %v", err)
	}

	err = json.Unmarshal(bytes, &results)
	if err != nil {
		metrics.RecordError("artifactory_annotator", "unable to Unmarshal ScanResults")
		return nil, fmt.Errorf("unable to Unmarshal ScanResults from url %s: %v", ia.scanResultsURL, err)
	}

	return &results, nil
}

func (ia *ArtifactoryAnnotator) addAnnotationsToImages(results perceptorapi.ScanResults) {
	regs := 0

	for _, registry := range ia.registryAuths {
		imgs := 0
		for _, image := range results.Images {

			baseURL := fmt.Sprintf("https://%s", registry.URL)
			cred, err := utils.PingArtifactoryServer(baseURL, registry.User, registry.Password)

			if err != nil {
				log.Warnf("Annotator: URL %s either not a valid Artifactory repository or incorrect credentials: %e", baseURL, err)
				break
			}

			regs = regs + 1
			repos := &ReposBySha{}
			// Look for SHA
			url := fmt.Sprintf("%s/artifactory/api/search/checksum?sha256=%s", baseURL, image.Sha)
			err = utils.GetResourceOfType(url, cred, "", repos)
			if err != nil {
				log.Errorf("Error in getting docker repo: %e", err)
				break
			}

			imgs = imgs + 1
			log.Debugf("Total Repos for image %s in artifactory: %d", image.Repository, len(repos.Results))
			for _, repo := range repos.Results {
				uri := strings.Replace(repo.URI, "/manifest.json", "", -1)
				ia.AnnotateImage(uri, &image, cred)
			}

		}

		log.Infof("Total images in Artifactory with URL %s: %d", registry.URL, imgs)
	}

	log.Infof("Total valid Artifactory Registries: %d", regs)
}

// AnnotateImage takes the specific Artifactory URL and applies the properties/annotations given by BD
func (ia *ArtifactoryAnnotator) AnnotateImage(uri string, im *perceptorapi.ScannedImage, cred *utils.RegistryAuth) {
	log.Infof("Annotating image in artifactory %s with URI %s", im.Repository, uri)
	url := fmt.Sprintf("%s?properties=%s=%s;%s=%d;%s=%d;%s=%s;", uri, bdSt, im.OverallStatus, bdVuln, im.Vulnerabilities, bdPolicy, im.PolicyViolations, bdComp, im.ComponentsURL)
	req, err := http.NewRequest(http.MethodPut, url, nil)
	if err != nil {
		log.Errorf("Error in creating put request %e", err)
	}
	req.SetBasicAuth(cred.User, cred.Password)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Errorf("Error in sending request %e", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		log.Errorf("Server is supposed to return status code %d given status code %d", http.StatusNoContent, resp.StatusCode)
	} else {
		log.Infof("Properties successfully added/updated for %s:%s", im.Repository, im.Tag)
	}

}
