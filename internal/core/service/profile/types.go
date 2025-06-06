package service

type ProfileServiceInterface interface {
	ProfilePersonServiceInterface
	ProfileProfessionServiceInterface
}

type ProfileAllService struct {
	ProfilePersonServiceInterface
	ProfileProfessionServiceInterface
}

func InicializeProfileAllService(person ProfilePersonServiceInterface, profession ProfileProfessionServiceInterface) (ProfileServiceInterface, error) {
	return &ProfileAllService{
		ProfilePersonServiceInterface:     person,
		ProfileProfessionServiceInterface: profession,
	}, nil
}
