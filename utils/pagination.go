package utils

import (
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func Paginate(c *fiber.Ctx, db *gorm.DB, out interface{}) error {
    
    fmt.Println(c.Query("skip"))

    // get the total count of records
    var count int64
    if err := db.Model(out).Count(&count).Error; err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "status":  "error",
            "message": "Could not retrieve data",
        })
    }

    // get the limit and skip parameters from the query string
    limit, err := strconv.Atoi(c.Query("limit", "10"))
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "status":  "error",
            "message": "Invalid limit parameter",
        })
    }

    skip, err := strconv.ParseInt(c.Query("skip", "0"), 10, 64)
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "status":  "error",
            "message": "Invalid skip parameter",
        })
    }

    // make sure skip is within range
    // if skip >= count {
    //     skip = count - 1
    // }
	//remaining := count - skip


    // query the records with the limit and skip parameters
	err = db.Limit(int(limit)).Offset(int(skip)).Find(out).Error
	
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "status":  "error",
            "message": "Could not retrieve data",
        })
    }

    return c.JSON(fiber.Map{
        "status": "success",
        "data":   out,
        "meta": fiber.Map{
            "total": count,
            "limit": limit,
            "skip":  skip,
        },
    })
}

func PaginateShort(c *fiber.Ctx, db *gorm.DB, out interface{}) error {
    
      // get the limit and skip parameters from the query string
    limit, err := strconv.Atoi(c.Query("limit", "10"))
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "status":  "error",
            "message": "Invalid limit parameter",
        })
    }

    skip, err := strconv.Atoi(c.Query("skip", "0"))
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "status":  "error",
            "message": "Invalid skip parameter",
        })
    }

    // query the records with the limit and skip parameters
    err = db.Limit(limit).Offset(skip).Find(out).Error
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "status":  "error",
            "message": "Could not retrieve data",
        })
    }

    // get the total count of records
    var count int64
    if err := db.Model(out).Count(&count).Error; err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "status":  "error",
            "message": "Could not retrieve data",
        })
    }

    return c.JSON(fiber.Map{
        "status": "success",
        "data":   out,
        "meta": fiber.Map{
            "total": count,
            "limit": limit,
            "skip":  skip,
        },
    })
} 