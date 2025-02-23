package httpserver

type HttpServer struct {
	largestPictureService MarsAPILargestPictureService
}

func NewHttpServer(largestPictureService MarsAPILargestPictureService) HttpServer {
	return HttpServer{
		largestPictureService: largestPictureService,
	}
}
