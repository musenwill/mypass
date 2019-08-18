package manager

func InitStore(gitUrl string) error {
	if !pathExists(passdir()) {
		err := createDir(passdir())
		if err != nil {
			return err
		}
	}
	if !pathExists(passfile()) {
		err := createFile(passfile())
		if err != nil {
			return err
		}
	}
	if !pathExists(configfile()) {
		err := createFile(configfile())
		if err != nil {
			return err
		}
	}

	conf := &config{Git: gitUrl}
	return saveConf(conf, configfile())
}

func Groups() ([]string, error) {
	return nil, nil
}

func Titles() ([]string, error) {
	return nil, nil
}

func Filter(grouplike, titleLike string) ([][]string, error) {
	return nil, nil
}

func Delete(group, title string) error {
	return nil
}

func Put(group, title, describe string) error {
	return nil
}

func Get(title string, print bool) ([]string, error) {

	return nil, nil
}

func History(title string) ([][]string, error) {
	return nil, nil
}
