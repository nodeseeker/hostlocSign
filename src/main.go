package main

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	SleepTime int    `json:"sleep_time"`
	UserAgent string `json:"user_agent"`
	Telegram  struct {
		Enable bool   `json:"enable"`
		Token  string `json:"token"`
		ChatID string `json:"chat_id"`
	} `json:"telegram"`
	Accounts []struct {
		Username string `json:"username"`
		Password string `json:"password"`
	} `json:"accounts"`
}

// get config from config.json
func getConfig() (*Config, error) {
	exePath, err := os.Executable()
	if err != nil {
		return nil, err
	}

	configPath := filepath.Join(filepath.Dir(exePath), "config.json")

	// fmt.Printf("config path: %s\n", configPath)

	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	config := &Config{}
	err = json.NewDecoder(file).Decode(config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

// write log to log.txt
func writeLog(log string) error {
	exePath, err := os.Executable()
	if err != nil {
		return err
	}

	logPath := filepath.Join(filepath.Dir(exePath), "scores.log")

	// fmt.Printf("log path: %s\n", logPath)

	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(log + "\n")
	if err != nil {
		return err
	}

	return nil
}

func sendMsg(token string, chatID string, msg string) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage?chat_id=%s&text=%s", token, chatID, msg)

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %s", resp.Status)
	}

	return nil
}

func login(username string, password string, useragent string) (*http.Client, error) {
	headers := http.Header{
		"User-Agent": []string{useragent},
		"Origin":     []string{"https://hostloc.com"},
		"Referer":    []string{"https://hostloc.com/forum.php"},
	}

	loginURL := "https://hostloc.com/member.php?mod=logging&action=login&loginsubmit=yes&infloat=yes&lssubmit=yes&inajax=1"
	loginData := url.Values{
		"fastloginfield": []string{"username"},
		"username":       []string{username},
		"password":       []string{password},
		"quickforward":   []string{"yes"},
		"handlekey":      []string{"ls"},
	}

	jar, _ := cookiejar.New(nil)
	client := &http.Client{
		Jar: jar,
	}

	req, err := http.NewRequest("POST", loginURL, strings.NewReader(loginData.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header = headers
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %s", resp.Status)
	}

	return client, nil
}

func checkScores(client *http.Client) (int, error) {
	res, err := client.Get("https://hostloc.com/forum.php")
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("unexpected status: %s", res.Status)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return 0, err
	}

	//get score
	re := regexp.MustCompile(`积分: (\d+)`)
	matches := re.FindStringSubmatch(string(body))

	// check if the value is larger than 0
	if len(matches) < 2 {
		return 0, fmt.Errorf("unexpected score: %s", matches[1])
	}

	score, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0, err
	}

	return score, nil
}

func getScore(client *http.Client, timesleep int) error {
	// generate 15 urls randomly between 10000 and 50000
	urls := make([]string, 15)
	for i := 0; i < 15; i++ {
		urls[i] = fmt.Sprintf("https://hostloc.com/space-uid-%d.html", 10000+rand.Intn(50000))
	}

	for _, url := range urls {

		// fmt.Printf("GET %s\n", url)

		resp, err := client.Get(url)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("unexpected status: %s", resp.Status)
		}

		time.Sleep(time.Duration(timesleep) * time.Second)
	}

	return nil
}

func checkLogin(client *http.Client) (bool, error) {
	resp, err := client.Get("https://hostloc.com/forum.php")
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("unexpected status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}

	return strings.Contains(string(body), "我的空间"), nil
}

func main() {

	// get config
	config, err := getConfig()
	if err != nil {
		fmt.Println("get config error: ", err)

		// send message to telegram
		if config.Telegram.Enable {
			err = sendMsg(config.Telegram.Token, config.Telegram.ChatID, err.Error())
			if err != nil {
				fmt.Println("send message & get config error: ", err)
			}
		}

		return
	}

	for _, account := range config.Accounts {
		// fmt.Printf("username: %s, password: %s\n", account.Username, account.Password)

		username := account.Username
		password := account.Password

		client, err := login(username, password, config.UserAgent)
		if err != nil {
			// print username + "login error" + error
			fmt.Println(username, "login error", err)

			// send message to telegram
			if config.Telegram.Enable {
				err = sendMsg(config.Telegram.Token, config.Telegram.ChatID, err.Error())
				if err != nil {
					fmt.Println("send message & login error: ", err)
				}
			}

			return
		}

		ok, err := checkLogin(client)
		if err != nil {
			fmt.Println("check login error: ", err)

			// send message to telegram
			if config.Telegram.Enable {
				err = sendMsg(config.Telegram.Token, config.Telegram.ChatID, err.Error())
				if err != nil {
					fmt.Println("send message & check login error: ", err)
				}
			}

			return
		}

		if ok {

			// fmt.Println("登录成功")

			initalScore, err := checkScores(client)
			if err != nil {
				fmt.Println("check score error: ", err)

				// send message to telegram
				if config.Telegram.Enable {
					err = sendMsg(config.Telegram.Token, config.Telegram.ChatID, err.Error())
					if err != nil {
						fmt.Println("send message & check score error: ", err)
					}
				}

				return
			}
			//fmt.Println("初始积分：", initalScore)

			err = getScore(client, config.SleepTime)
			if err != nil {
				fmt.Println("inital get score error: ", err)

				// send message to telegram
				if config.Telegram.Enable {
					err = sendMsg(config.Telegram.Token, config.Telegram.ChatID, err.Error())
					if err != nil {
						fmt.Println("send message & inital get score error: ", err)
					}
				}

				return
			}

			finalScore, err := checkScores(client)
			if err != nil {
				fmt.Println("final check score error: ", err)

				// send message to telegram
				if config.Telegram.Enable {
					err = sendMsg(config.Telegram.Token, config.Telegram.ChatID, err.Error())
					if err != nil {
						fmt.Println("send message & final check score error: ", err)
					}
				}

				return
			}
			// fmt.Println("当前积分：", finalScore)

			// write log, date + time + username + initalScore + finalScore
			log := fmt.Sprintf("%s,%s,%d,%d", time.Now().Format("2006-01-02 15:04:05"), username, initalScore, finalScore)
			err = writeLog(log)
			if err != nil {
				fmt.Println("write log error: ", err)

				// send message to telegram
				if config.Telegram.Enable {
					err = sendMsg(config.Telegram.Token, config.Telegram.ChatID, err.Error())
					if err != nil {
						fmt.Println("send message & write log error: ", err)
					}
				}
			}

		} else {
			dateTime := time.Now().Format("2006-01-02 15:04:05")
			fmt.Printf("%s %s 签到登录失败\n", dateTime, username)
			// send message to telegram
			if config.Telegram.Enable {
				msg := fmt.Sprintf("%s %s 签到登录失败", dateTime, username)
				err = sendMsg(config.Telegram.Token, config.Telegram.ChatID, msg)
				if err != nil {
					fmt.Println("send message & login error: ", err)
				}
			}
		}

	}

}
