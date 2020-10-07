package service

func InitializeRoutes(a *ApiService) {
	//a.router.GET("/list", a.List)
	a.router.POST("/create", a.Create)
	a.router.PATCH("/update", a.Update)
}
