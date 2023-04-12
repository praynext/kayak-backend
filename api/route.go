package api

import "kayak-backend/global"

func InitRoute() {
	global.Router.GET("/ping", Ping)
	global.Router.GET("/logout", Logout)
	global.Router.POST("/login", Login)
	global.Router.POST("/register", Register)
	global.Router.POST("/reset-password", ResetPassword)

	user := global.Router.Group("/user")
	user.Use(global.CheckAuth)
	user.GET("/info", GetUserInfo)
	user.GET("/info/:user_id", GetUserInfoById)
	user.PUT("/update", UpdateUserInfo)
	user.GET("/wrong_record", GetUserWrongRecords)

	upload := global.Router.Group("/upload")
	upload.Use(global.CheckAuth)
	upload.POST("/public", UploadPublicFile)
	upload.POST("/avatar", UploadAvatar)

	note := global.Router.Group("/note")
	note.Use(global.CheckAuth)
	global.Router.GET("/note/all", GetNotes)
	note.POST("/create", CreateNote)
	note.PUT("/update", UpdateNote)
	note.DELETE("/delete/:id", DeleteNote)
	note.POST("/like/:id", LikeNote)
	note.POST("/unlike/:id", UnlikeNote)
	note.POST("/favorite/:id", FavoriteNote)
	note.DELETE("/unfavorite/:id", UnfavoriteNote)

	wrongRecord := global.Router.Group("/wrong_record")
	wrongRecord.Use(global.CheckAuth)
	wrongRecord.POST("/create/:id", CreateWrongRecord)
	wrongRecord.DELETE("/delete/:id", DeleteWrongRecord)

	problem := global.Router.Group("/problem")
	problem.Use(global.CheckAuth)
	problem.DELETE("/unfavorite/:id", RemoveProblemFromFavorite)
	problem.POST("/favorite/:id", AddProblemToFavorite)

	choiceProblem := problem.Group("/choice")
	global.Router.GET("/problem/choice/all", GetChoiceProblems)
	choiceProblem.POST("/create", CreateChoiceProblem)
	choiceProblem.PUT("/update", UpdateChoiceProblem)
	choiceProblem.DELETE("/delete/:id", DeleteChoiceProblem)
	choiceProblem.GET("/answer/:id", GetChoiceProblemAnswer)

	blankProblem := problem.Group("/blank")
	global.Router.GET("/problem/blank/all", GetBlankProblems)
	blankProblem.POST("/create", CreateBlankProblem)
	blankProblem.PUT("/update", UpdateBlankProblem)
	blankProblem.DELETE("/delete/:id", DeleteBlankProblem)
	blankProblem.GET("/answer/:id", GetBlankProblemAnswer)

	judgeProblem := problem.Group("/judge")
	global.Router.GET("/problem/judge/all", GetJudgeProblems)
	judgeProblem.POST("/create", CreateJudgeProblem)
	judgeProblem.PUT("/update", UpdateJudgeProblem)
	judgeProblem.DELETE("/delete/:id", DeleteJudgeProblem)
	judgeProblem.GET("/answer/:id", GetJudgeProblemAnswer)

	problemSet := global.Router.Group("/problem_set")
	problemSet.Use(global.CheckAuth)
	global.Router.GET("/problem_set/all", GetProblemSets)
	problemSet.POST("/create", CreateProblemSet)
	problemSet.DELETE("/delete/:id", DeleteProblemSet)
	problemSet.GET("/:id/all_problem", GetProblemsInProblemSet)
	problemSet.POST("/:id/add", AddProblemToProblemSet)
	problemSet.DELETE("/:id/remove", RemoveProblemFromProblemSet)
	problemSet.POST("/favorite/:id", AddProblemSetToFavorite)
	problemSet.DELETE("/unfavorite/:id", RemoveProblemSetFromFavorite)

	noteReview := global.Router.Group("/note_review")
	noteReview.Use(global.CheckAuth)
	noteReview.POST("/add", AddNoteReview)
	noteReview.DELETE("/remove/:id", RemoveNoteReview)
	noteReview.GET("/get", GetNoteReviews)

	// Deprecated
	global.Router.GET("/problem/blank/:id", GetBlankProblem)
	global.Router.GET("/problem/choice/:id", GetChoiceProblem)
	problem.GET("/:id/problem_set", GetProblemSetContainsProblem)
	user.GET("/favorite/problem", GetUserFavoriteProblems)
	user.GET("/favorite/problem_set", GetUserFavoriteProblemSets)
	user.GET("/favorite/note", GetUserFavoriteNotes)
	user.GET("/problem/choice", GetUserChoiceProblems)
	user.GET("/problem/blank", GetUserBlankProblems)
	user.GET("/problem_set", GetUserProblemSets)
	user.GET("/note", GetUserNotes)
}
