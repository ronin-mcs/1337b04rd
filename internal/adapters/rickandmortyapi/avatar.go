package rickandmortyapi

type AvatarFromAPI struct {
	request string

}

func (a *AvatarFromAPI) NewAvatarFromAPI(characterName string) *AvatarFromAPI {
	return &AvatarFromAPI{
		request: "https://rickandmortyapi.com/api/character/?name=" + characterName,
	}
}

логика :

У каждого поста есть список anon у которых есть есть ссылка на их аватары 
