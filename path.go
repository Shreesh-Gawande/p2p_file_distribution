package main


func DefaultPathTransformFunc(key string) Path {
	return Path{
		Pathname: key,
		Filename: key,
	}
}