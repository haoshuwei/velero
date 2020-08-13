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
	"github.com/joho/godotenv"
	"github.com/pkg/errors"
	"os"
)

const (
	// AlibabaCloud specific environment variable
	alibabaCloudAccessKeyID     = "ALIBABA_CLOUD_ACCESS_KEY_ID"
	alibabaCloudAccessKeySecret = "ALIBABA_CLOUD_ACCESS_KEY_SECRET"
)

// getAlibabaCloudResticEnvVars gets the environment variables that restic
// relies on (ALIBABA_CLOUD_ACCESS_KEY_ID and ALIBABA_CLOUD_ACCESS_KEY_SECRET)
// based on info in the provided object storage location config map.
func getAlibabaCloudResticEnvVars(config map[string]string) (map[string]string, error) {
	result := make(map[string]string)

	// load environment vars from $ALIBABA_CLOUD_CREDENTIALS_FILE
	if err := loadAlibabaCloudEnv(); err != nil {
		return nil, err
	}

	// get ALIBABA_CLOUD_ACCESS_KEY_ID and ALIBABA_CLOUD_ACCESS_KEY_SECRET
	result[alibabaCloudAccessKeyID] = os.Getenv("ALIBABA_CLOUD_ACCESS_KEY_ID")
	result[alibabaCloudAccessKeySecret] = os.Getenv("ALIBABA_CLOUD_ACCESS_KEY_SECRET")

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