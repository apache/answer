/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package cli

import (
	"bufio"
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"os"
	"runtime"
	"strings"

	"github.com/apache/answer/internal/base/conf"
	"github.com/apache/answer/internal/base/data"
	"github.com/apache/answer/internal/base/path"
	"github.com/apache/answer/internal/repo/api_key"
	"github.com/apache/answer/internal/repo/auth"
	"github.com/apache/answer/internal/repo/user"
	authService "github.com/apache/answer/internal/service/auth"
	"github.com/apache/answer/pkg/checker"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/term"
	_ "modernc.org/sqlite"
	"xorm.io/xorm"
)

const (
	charsetLower                = "abcdefghijklmnopqrstuvwxyz"
	charsetUpper                = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	charsetDigits               = "0123456789"
	charsetSpecial              = "!@#$%^&*~?_-"
	maxRetries                  = 10
	defaultRandomPasswordLength = 12
)

var charset = []string{
	charsetLower,
	charsetUpper,
	charsetDigits,
	charsetSpecial,
}

type ResetPasswordOptions struct {
	Email    string
	Password string
}

func ResetPassword(ctx context.Context, dataDirPath string, opts *ResetPasswordOptions) error {
	path.FormatAllPath(dataDirPath)

	config, err := conf.ReadConfig(path.GetConfigFilePath())
	if err != nil {
		return fmt.Errorf("read config file failed: %w", err)
	}

	db, err := initDatabase(config.Data.Database.Driver, config.Data.Database.Connection)
	if err != nil {
		return fmt.Errorf("connect database failed: %w", err)
	}
	defer func() {
		_ = db.Close()
	}()

	cache, cacheCleanup, err := data.NewCache(config.Data.Cache)
	if err != nil {
		return fmt.Errorf("initialize cache failed: %w", err)
	}
	defer cacheCleanup()

	dataData, dataCleanup, err := data.NewData(db, cache)
	if err != nil {
		return fmt.Errorf("initialize data layer failed: %w", err)
	}
	defer dataCleanup()

	userRepo := user.NewUserRepo(dataData)
	authRepo := auth.NewAuthRepo(dataData)
	apiKeyRepo := api_key.NewAPIKeyRepo(dataData)
	authSvc := authService.NewAuthService(authRepo, apiKeyRepo)

	email := strings.TrimSpace(opts.Email)
	if email == "" {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Please input user email: ")
		emailInput, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("read email input failed: %w", err)
		}
		email = strings.TrimSpace(emailInput)
	}

	userInfo, exist, err := userRepo.GetByEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("query user failed: %w", err)
	}
	if !exist {
		return fmt.Errorf("user not found: %s", email)
	}

	fmt.Printf("You are going to reset password for user: %s\n", email)

	password := strings.TrimSpace(opts.Password)

	if password != "" {
		printWarning("Passing password via command line may be recorded in shell history")
		if err := checker.CheckPassword(password); err != nil {
			return fmt.Errorf("password validation failed: %w", err)
		}
	} else {
		password, err = promptForPassword()
		if err != nil {
			return fmt.Errorf("password input failed: %w", err)
		}
	}

	if !confirmAction(fmt.Sprintf("This will reset password for user '[%s]%s'. Continue?", userInfo.DisplayName, email)) {
		fmt.Println("Operation cancelled")
		return nil
	}

	hashPwd, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("encrypt password failed: %w", err)
	}

	if err = userRepo.UpdatePass(ctx, userInfo.ID, string(hashPwd)); err != nil {
		return fmt.Errorf("update password failed: %w", err)
	}

	authSvc.RemoveUserAllTokens(ctx, userInfo.ID)

	fmt.Printf("Password has been successfully updated for user: %s\n", email)
	fmt.Println("All login sessions have been cleared")

	return nil
}

// promptForPassword prompts for a password
func promptForPassword() (string, error) {
	for {
		input, err := getPasswordInput("Please input new password (empty to generate random password): ")
		if err != nil {
			return "", err
		}

		if input == "" {
			password, err := generateRandomPasswordWithRetry()
			if err != nil {
				return "", fmt.Errorf("generate random password failed: %w", err)
			}
			fmt.Printf("Generated random password: %s\n", password)
			fmt.Println("Please save this password in a secure location")
			return password, nil
		}

		if err := checker.CheckPassword(input); err != nil {
			fmt.Printf("Password validation failed: %v\n", err)
			fmt.Println("Please try again")
			continue
		}

		confirmPwd, err := getPasswordInput("Please confirm new password: ")
		if err != nil {
			return "", err
		}

		if input != confirmPwd {
			fmt.Println("Passwords do not match, please try again")
			continue
		}

		return input, nil
	}
}

func generateRandomPasswordWithRetry() (string, error) {
	var password string
	var err error

	for range maxRetries {
		password, err = generateRandomPassword(defaultRandomPasswordLength)
		if err != nil {
			continue
		}
		if err := checker.CheckPassword(password); err == nil {
			return password, nil
		}
	}

	if err != nil {
		return "", err
	}
	return "", fmt.Errorf("failed to generate valid password after %d retries", maxRetries)
}

func getPasswordInput(prompt string) (string, error) {
	fmt.Print(prompt)
	password, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return "", err
	}
	fmt.Println()
	return string(password), nil
}

func generateRandomPassword(length int) (string, error) {
	if length < len(charset) {
		return "", fmt.Errorf("password length must be at least %d", len(charset))
	}

	bytes := make([]byte, length)
	for i, charsetItem := range charset {
		charIndex, err := rand.Int(rand.Reader, big.NewInt(int64(len(charsetItem))))
		if err != nil {
			return "", err
		}
		bytes[i] = charsetItem[charIndex.Int64()]
	}

	fullCharset := strings.Join(charset, "")
	for i := len(charset); i < length; i++ {
		charIndex, err := rand.Int(rand.Reader, big.NewInt(int64(len(fullCharset))))
		if err != nil {
			return "", err
		}
		bytes[i] = fullCharset[charIndex.Int64()]
	}

	for i := len(bytes) - 1; i > 0; i-- {
		j, err := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		if err != nil {
			return "", err
		}
		bytes[i], bytes[j.Int64()] = bytes[j.Int64()], bytes[i]
	}

	return string(bytes), nil
}

func initDatabase(driver, connection string) (*xorm.Engine, error) {
	dataConf := &data.Database{Driver: driver, Connection: connection}
	if !CheckDBConnection(dataConf) {
		return nil, fmt.Errorf("database connection check failed")
	}

	engine, err := data.NewDB(false, dataConf)
	if err != nil {
		return nil, err
	}

	return engine, nil
}

func printWarning(msg string) {
	if runtime.GOOS == "windows" {
		fmt.Printf("[WARNING] %s\n", msg)
	} else {
		fmt.Printf("\033[31m[WARNING] %s\033[0m\n", msg)
	}
}

func confirmAction(prompt string) bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s [y/N]: ", prompt)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}
	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes"
}
