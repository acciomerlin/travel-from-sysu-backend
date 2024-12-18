package utils

//工具文件

import (
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"log"
	"time"
	"travel-from-sysu-backend/global"
	"travel-from-sysu-backend/models"
)

func HashPwd(pwd string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
	return string(hash), err
}

func GenerateJWT(username string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": username,
		"exp":      time.Now().Add(time.Hour * 72).Unix(),
	})
	signedToken, err := token.SignedString([]byte("secret"))
	return "Bearer " + signedToken, err
}

func CheckPwd(hashedPwd, plainPwd string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPwd), []byte(plainPwd))
}

// 归一化函数：将值归一化到 0 - 100 区间
func normalize(value, min, max int64) int64 {
	if max == min {
		return 0
	}
	return (value - min) * 100 / (max - min)
}

// 热度计算公式（计算时会归一化）
func calculateScore(note models.Note, minLikes, maxLikes, minCollects, maxCollects, minComments, maxComments, minTimestamp, maxTimestamp int64) float64 {
	// 归一化各项指标
	normalizedLikes := normalize(int64(note.LikeCounts), minLikes, maxLikes)
	normalizedCollects := normalize(int64(note.CollectCounts), minCollects, maxCollects)
	normalizedComments := normalize(int64(note.CommentCounts), minComments, maxComments)
	normalizedUpdateTime := normalize(note.NoteUpdateTime, minTimestamp, maxTimestamp)

	// 假设计算热度时，采用以下权重：收藏数 40%，点赞数 30%，评论数 20%，更新时间 10%
	return 0.4*float64(normalizedCollects) + 0.3*float64(normalizedLikes) +
		0.2*float64(normalizedComments) + 0.1*float64(normalizedUpdateTime)
}

// 每分钟计算一次热度并更新到数据库
func UpdateHotRecommendations() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		// 获取所有笔记数据
		var notes []models.Note
		if err := global.Db.Find(&notes).Error; err != nil {
			log.Printf("Failed to fetch notes for hot recommendations: %v", err)
			continue
		}

		// 计算最大最小值
		var maxCollects, minCollects, maxComments, minComments uint
		var maxLikes, minLikes int
		var maxTimestamp, minTimestamp int64

		// 获取最大值和最小值
		for _, note := range notes {
			if note.LikeCounts > maxLikes {
				maxLikes = note.LikeCounts
			}
			if note.LikeCounts < minLikes {
				minLikes = note.LikeCounts
			}
			if note.CollectCounts > maxCollects {
				maxCollects = note.CollectCounts
			}
			if note.CollectCounts < minCollects {
				minCollects = note.CollectCounts
			}
			if note.CommentCounts > maxComments {
				maxComments = note.CommentCounts
			}
			if note.CommentCounts < minComments {
				minComments = note.CommentCounts
			}
			if note.NoteUpdateTime > maxTimestamp {
				maxTimestamp = note.NoteUpdateTime
			}
			if note.NoteUpdateTime < minTimestamp {
				minTimestamp = note.NoteUpdateTime
			}
		}

		// 计算每个笔记的热度并更新到数据库
		for _, note := range notes {
			// 计算归一化后的热度
			score := calculateScore(note, int64(minLikes), int64(maxLikes), int64(minCollects), int64(maxCollects), int64(minComments), int64(maxComments), minTimestamp, maxTimestamp)

			// 更新笔记的热度分数
			if err := global.Db.Model(&note).Update("score", score).Error; err != nil {
				log.Printf("Failed to update score for note %d: %v", note.NoteID, err)
			}
		}

		log.Println("Hot recommendations updated")
	}
}
