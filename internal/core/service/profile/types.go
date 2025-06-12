package profile

type ProfileServiceInterface interface {
	ProfilePersonServiceInterface
	ProfileProfessionServiceInterface
	ProfileGoalsServiceInterface
}

type ProfileAllService struct {
	ProfilePersonServiceInterface
	ProfileProfessionServiceInterface
	ProfileGoalsServiceInterface
}

func InicializeProfileAllService(person ProfilePersonServiceInterface, profession ProfileProfessionServiceInterface, goals ProfileGoalsServiceInterface) (ProfileServiceInterface, error) {
	return &ProfileAllService{
		ProfilePersonServiceInterface:     person,
		ProfileProfessionServiceInterface: profession,
		ProfileGoalsServiceInterface:      goals,
	}, nil
}
