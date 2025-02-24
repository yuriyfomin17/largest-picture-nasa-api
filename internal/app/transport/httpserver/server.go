package httpserver

type HttpServer struct {
	largestPictureService MarsApiLargestPictureService
}

func NewHttpServer(largestPictureService MarsApiLargestPictureService) HttpServer {
	return HttpServer{
		largestPictureService: largestPictureService,
	}
}
