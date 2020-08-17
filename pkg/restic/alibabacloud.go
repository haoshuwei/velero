/*
Copyright 2019 the Velero contributors.

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

package restic

import (
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

const (
	// AlibabaCloud specific environment variable
	alibabaCloudAccessKeyID     = "ALIBABA_CLOUD_ACCESS_KEY_ID"
	alibabaCloudAccessKeySecret = "ALIBABA_CLOUD_ACCESS_KEY_SECRET"
	alibabaCloudStsToken        = "ALIBABA_CLOUD_STS_TOKEN"
	metadataURL                 = "http://100.100.100.200/latest/meta-data/"
)

// RoleAuth define STS Token Response
type RoleAuth struct {
	AccessKeyID     string
	AccessKeySecret string
	Expiration      time.Time
	SecurityToken   string
	LastUpdated     time.Time
	Code            string
}

// getAlibabaCloudResticEnvVars gets the environment variables that restic
// relies on (ALIBABA_CLOUD_ACCESS_KEY_ID and ALIBABA_CLOUD_ACCESS_KEY_SECRET)
// based on info in the provided object storage location config map.
func getAlibabaCloudResticEnvVars(config map[string]string) (map[string]string, error) {
	result := make(map[string]string)

	accessKeyID := ""
	accessKeySecret := ""
	stsToken := ""

	veleroForAck := os.Getenv("VELERO_FOR_ACK")
	isHybrid := os.Getenv("IS_HYBRID")

	if veleroForAck == "true" {
		if isHybrid == "true" {
			accessKeyID = os.Getenv("ALIBABA_CLOUD_ACCESS_KEY_ID")
			accessKeySecret = os.Getenv("ALIBABA_CLOUD_ACCESS_KEY_SECRET")
			if len(accessKeyID) == 0 {
				return nil, errors.Errorf("IS_HYBRID set to true, but ALIBABA_CLOUD_ACCESS_KEY_ID environment variable is not set")
			}
			if len(accessKeySecret) == 0 {
				return nil, errors.Errorf("IS_HYBRID set to true, but ALIBABA_CLOUD_ACCESS_KEY_SECRET environment variable is not set")
			}
		} else {
			ramRole, err := getRamRole()
			if err != nil {
				return nil, errors.Errorf("Failed to get ram role with err: %v", err)
			}

			accessKeyID, accessKeySecret, stsToken, err = getSTSAK(ramRole)
			if err != nil {
				return nil, errors.Errorf("Failed to get sts token from ram role %s with err: %v", ramRole, err)
			}
		}
	} else {
		// load environment vars from $ALIBABA_CLOUD_CREDENTIALS_FILE
		if err := loadAlibabaCloudEnv(); err != nil {
			return nil, err
		}
	}

	// get ALIBABA_CLOUD_ACCESS_KEY_ID and ALIBABA_CLOUD_ACCESS_KEY_SECRET
	result[alibabaCloudAccessKeyID] = accessKeyID
	result[alibabaCloudAccessKeySecret] = accessKeySecret
	result[alibabaCloudStsToken] = stsToken

	return result, nil
}

// load environment vars from $ALIBABA_CLOUD_CREDENTIALS_FILE, if it exists
func loadAlibabaCloudEnv() error {

	envFile := os.Getenv("ALIBABA_CLOUD_CREDENTIALS_FILE")
	if envFile == "" {
		return nil
	}

	if err := godotenv.Overload(envFile); err != nil {
		return errors.Wrapf(err, "error loading environment from ALIBABA_CLOUD_CREDENTIALS_FILE (%s)", envFile)
	}

	return nil
}

// getRamRole return ramrole name
func getRamRole() (string, error) {
	subpath := "ram/security-credentials/"
	roleName, err := getMetaData(subpath)
	if err != nil {
		return "", err
	}
	return roleName, nil
}

//getMetaData get metadata from ecs meta-server
func getMetaData(resource string) (string, error) {
	resp, err := http.Get(metadataURL + resource)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

//getSTSAK return AccessKeyID, AccessKeySecret and SecurityToken
func getSTSAK(ramrole string) (string, string, string, error) {
	// AliyunCSVeleroRole
	roleAuth := RoleAuth{}
	ramRoleURL := fmt.Sprintf("ram/security-credentials/%s", ramrole)
	roleInfo, err := getMetaData(ramRoleURL)
	if err != nil {
		return "", "", "", err
	}

	err = json.Unmarshal([]byte(roleInfo), &roleAuth)
	if err != nil {
		return "", "", "", err
	}
	return roleAuth.AccessKeyID, roleAuth.AccessKeySecret, roleAuth.SecurityToken, nil
}