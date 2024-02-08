package api

import (
	"context"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"

	"hyperpage/controllers"
	"hyperpage/initializers"
	"hyperpage/middleware"
)

func Register(micro *fiber.App) {
	micro.Route("/settings", func(router fiber.Router) {
		router.Get("/langs", controllers.Langs)
		router.Post("/addlang", middleware.DeserializeUser, middleware.CheckRole([]string{"admin"}), controllers.AddLang)
		router.Delete("/deletelang/:id", middleware.DeserializeUser, middleware.CheckRole([]string{"admin"}), controllers.DeleteLang)
		router.Patch("/updatelang/:id", middleware.DeserializeUser, middleware.CheckRole([]string{"admin"}), controllers.UpdateLang)
	})

	micro.Route("/devices", func(router fiber.Router) {
		router.Post("/ios", controllers.CreateDevice)
	})

	micro.Route("/auth", func(router fiber.Router) {
		router.Post("/register", controllers.SignUpUser)

		router.Post("/login", controllers.SignInUser)
		router.Post("/forgotpassword", controllers.ForgotPassword)

		router.Patch("/resetpassword/:resetToken", controllers.ResetPassword)

		router.Get("/verifyemail/:verificationCode", controllers.VerifyEmail)
		router.Get("/logout", controllers.LogoutUser)
		router.Get("/refresh/:refreshToken", controllers.RefreshAccessToken)
		router.Post("/checkTokenExp", controllers.CheckTokenExp)
	})

	micro.Route("/followers", func(router fiber.Router) {
		router.Post("/scribe", middleware.DeserializeUser, controllers.Scribe)
		router.Post("/unscribe", middleware.DeserializeUser, controllers.Unscribe)
		router.Get("/get", middleware.DeserializeUser, controllers.GetFollowers)
	})

	micro.Route("/domains", func(router fiber.Router) {
		router.Get("/get", controllers.GetDomain)
	})

	micro.Route("/site", func(router fiber.Router) {
		router.Post("/update", middleware.DeserializeUser, controllers.UpdateSite)
		router.Get("/get", middleware.DeserializeUser, controllers.GetSite)
	})

	micro.Route("/users", func(router fiber.Router) {
		router.Get("/myTime", controllers.MyTime)
		router.Post("/deletme", middleware.DeserializeUser, controllers.DeleteUserWithRelations)
		router.Post("/setvip", middleware.DeserializeUser, controllers.SetVipUser)

		router.Post("/sendrequestcall", controllers.SendBotCallRequest)
		// router.Get("/me", middleware.DeserializeUser, controllers.GetMe)
		router.Get("/me", func(c *fiber.Ctx) error {
			// Capture the language from the URL, headers, or any other source.
			language := c.Query("language") // Example: ?language=en
			if language == "" {
				language = "en"
			}
			// Set the language in the context for middleware.
			c.Locals("language", language)

			// Call the DeserializeUser middleware.
			return middleware.DeserializeUser(c)
		}, controllers.GetMe)
		router.Get("/getmefirst", middleware.DeserializeUser, controllers.GetMeFirst)
		router.Post("/addbalance", middleware.DeserializeUser, controllers.AddBalance)
		router.Post("/plan", middleware.DeserializeUser, controllers.Plan)
	})

	micro.Route("/billing", func(router fiber.Router) {
		router.Get("/transactions", middleware.DeserializeUser, controllers.GetTransactions)
	})

	micro.Route("/calls", func(router fiber.Router) {
		router.Post("/makecall", controllers.MakeCall)
		router.Post("/stopcall", controllers.StopCall)

	})

	micro.Route("/cities", func(router fiber.Router) {
		router.Get("/all", controllers.GetCities)
		router.Get("/query", controllers.GetName)
		router.Post("/create", middleware.DeserializeUser, middleware.CheckRole([]string{"admin"}), controllers.CreateCity)
		router.Delete("/remove/:id", middleware.DeserializeUser, middleware.CheckRole([]string{"admin"}), controllers.DeleteCity)
		router.Patch("/update/:id", middleware.DeserializeUser, middleware.CheckRole([]string{"admin"}), controllers.UpdateCity)
		router.Get("/get", middleware.DeserializeUser, middleware.CheckRole([]string{"admin"}), controllers.GetCityTranslation)
	})

	micro.Route("/citiestranslator", func(router fiber.Router) {
		router.Post("/create", middleware.DeserializeUser, middleware.CheckRole([]string{"admin"}), controllers.CreateCityTranslation)
		router.Delete("/remove", middleware.DeserializeUser, middleware.CheckRole([]string{"admin"}), controllers.DeleteCityTranslation)
		router.Patch("/update", middleware.DeserializeUser, middleware.CheckRole([]string{"admin"}), controllers.UpdateCityTranslation)
	})

	micro.Route("/guilds", func(router fiber.Router) {
		router.Get("/all", controllers.GetGuilds)
		router.Get("/getAll", controllers.GetGuildsAll)
		router.Post("/create", middleware.DeserializeUser, middleware.CheckRole([]string{"admin"}), controllers.CreateGuild)
		router.Delete("/remove/:id", middleware.DeserializeUser, middleware.CheckRole([]string{"admin"}), controllers.DeleteGuild)
		router.Patch("/update/:id", middleware.DeserializeUser, middleware.CheckRole([]string{"admin"}), controllers.UpdateGuild)

		router.Get("/name", controllers.GetGuildName)
		router.Get("/namecustom", controllers.GetGuildNameA)

	})

	micro.Route("/guildstranslator", func(router fiber.Router) {
		router.Post("/create", middleware.DeserializeUser, middleware.CheckRole([]string{"admin"}), controllers.CreateGuildTranslation)
		router.Delete("/remove", middleware.DeserializeUser, middleware.CheckRole([]string{"admin"}), controllers.DeleteGuildTranslation)
		router.Patch("/update", middleware.DeserializeUser, middleware.CheckRole([]string{"admin"}), controllers.UpdateGuildTranslation)
	})

	micro.Route("/profile", func(router fiber.Router) {
		router.Get("/get", middleware.DeserializeUser, middleware.CheckRole([]string{"admin", "user", "vip"}), controllers.GetProfile)
		router.Patch("/save", middleware.DeserializeUser, middleware.CheckRole([]string{"admin", "user", "vip"}), controllers.UpdateProfile)
		router.Patch("/saveAdditional", middleware.DeserializeUser, middleware.CheckRole([]string{"admin", "user", "vip"}), controllers.UpdateProfileAdditional)

		router.Patch("/photos", middleware.DeserializeUser, middleware.CheckRole([]string{"admin", "user", "vip"}), controllers.UpdateProfilePhotos)
		router.Post("/documents", middleware.DeserializeUser, middleware.CheckRole([]string{"admin", "user", "vip"}), controllers.NewProfileDocuments)
		router.Patch("/documents", middleware.DeserializeUser, middleware.CheckRole([]string{"admin", "user", "vip"}), controllers.UpdateProfileDocuments)
		router.Delete("/documents/:id", middleware.DeserializeUser, middleware.CheckRole([]string{"admin", "user", "vip"}), controllers.DeleteProfileDocuments)

		router.Get("/getdocuments", middleware.DeserializeUser, middleware.CheckRole([]string{"admin", "user", "vip"}), controllers.GetDocuments)
	})

	micro.Route("/profiles", func(router fiber.Router) {
		router.Get("/get", controllers.GetAllProfile)
		router.Get("/get/:name", controllers.GetProfileGuest)
	})

	micro.Route("/payment", func(router fiber.Router) {
		router.Post("/invoice", middleware.DeserializeUser, controllers.CreateInvoice)
		router.Post("/pending", controllers.Pending)

	})

	micro.Route("/profilehashtags", func(router fiber.Router) {
		router.Post("/addhashtag", middleware.DeserializeUser, middleware.CheckRole([]string{"admin", "user", "vip"}), controllers.AddHashTagProfile)
		router.Get("/findTag", controllers.SearchHashTagProfile)
	})

	micro.Route("/blog", func(router fiber.Router) {
		router.Get("/list", middleware.DeserializeUser, middleware.CheckRole([]string{"admin", "user", "vip"}), controllers.GetAllBlogs)
		router.Post("/makearchive/:id", middleware.DeserializeUser, middleware.CheckRole([]string{"admin", "user", "vip"}), controllers.SendToArchive)
		router.Post("/search", middleware.DeserializeUser, controllers.SearchBlogByTitle)
		router.Post("/addblogtime", middleware.DeserializeUser, controllers.AddBlogTime)
		router.Post("/addhashtag", middleware.DeserializeUser, controllers.AddHashTag)
		router.Get("/findTag", controllers.SearchHashTag)

		router.Get("/allvotes/:id", controllers.GetAllVotes)
		router.Post("/addvote/:id", middleware.DeserializeUser, controllers.AddVote)

		router.Get("/getAllByUser/:id", controllers.GetAllByUser)

		router.Get("/listAll", controllers.GetAll)

		router.Get("/random", controllers.GetRandom)

		router.Get("/:id", controllers.GetBlogById)
		router.Post("/create", middleware.DeserializeUser, middleware.CheckRole([]string{"admin", "user", "vip"}), middleware.CheckProfileFilled(), controllers.CreateBlog)
		router.Post("/create/photos", middleware.DeserializeUser, controllers.CreateBlogPhoto)
		router.Get("/edit/:id", middleware.DeserializeUser, middleware.CheckRole([]string{"admin", "user", "vip"}), controllers.EditBlogGetId)
		router.Patch("/patch/:id", middleware.DeserializeUser, middleware.CheckRole([]string{"admin", "user", "vip"}), controllers.UpdateBlog)
		router.Delete("/delete/:id", middleware.DeserializeUser, middleware.CheckRole([]string{"admin", "user", "vip"}), controllers.DeleteBlog)
	})

	micro.Route("/files", func(router fiber.Router) {
		router.Post("/upload/file", middleware.DeserializeUser, middleware.CheckProfileFilled(), controllers.UploadPdf)
		router.Post("/upload", middleware.DeserializeUser, middleware.CheckProfileFilled(), controllers.UploadImage)
		router.Post("/upload/images", middleware.DeserializeUser, middleware.CheckProfileFilled(), controllers.UploadImages)

	})

	micro.Route("/server", func(router fiber.Router) {
		ctx := context.TODO()
		value, err := initializers.RedisClient.Get(ctx, "statusHealth").Result()

		router.Get("/healthchecker", func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusOK).JSON(fiber.Map{
				"status":  "success",
				"message": value,
			})
		})

		if err == redis.Nil {
			fmt.Println("key: statusHealth does not exist")
		} else if err != nil {
			panic(err)
		}
	})

	micro.Route("/managebot", func(router fiber.Router) {
		router.Post("/registerbot", controllers.SignUpBot)
		router.Post("/deletbots", controllers.DeleteAllBotUsersWithRelations)
	})

	micro.All("*", func(c *fiber.Ctx) error {
		path := c.Path()
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"status":  "fail",
			"message": fmt.Sprintf("Path: %v does not exists", path),
		})
	})
}
