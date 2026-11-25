package cmd

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/musenwill/mypass/data"
	"github.com/musenwill/mypass/errs"
	"github.com/musenwill/mypass/manager"

	"github.com/urfave/cli"
)

func oldPasswords(c *cli.Context) error {
	srv, err := loadStore(false)
	if err != nil {
		return err
	}

	now := time.Now()
	halfYearAgo := now.AddDate(0, -6, 0)

	result, err := srv.Olds(halfYearAgo)
	if err != nil {
		return err
	}

	if len(result) > 0 {
		fmt.Println("password of these accounts were updated 6 months ago, they may be in risk, suggest update them now")
		printRecords(result...)
	}

	return nil
}

func initStore(c *cli.Context) error {
	gitUrl := flagString(c, flagGit)
	srv := manager.New()
	return srv.Init(gitUrl)
}

func all(c *cli.Context) error {
	srv, err := load()
	if err != nil {
		return err
	}

	result, err := srv.All()
	if err != nil {
		return err
	}

	printRecords(result...)

	return nil
}

func groups(c *cli.Context) error {
	srv, err := load()
	if err != nil {
		return err
	}

	result, err := srv.Groups()
	if err != nil {
		return err
	}

	fmt.Printf("%v\n", result)

	return nil
}

func titles(c *cli.Context) error {
	srv, err := load()
	if err != nil {
		return err
	}

	result, err := srv.Titles()
	if err != nil {
		return err
	}

	fmt.Printf("%v\n", result)

	return nil
}

func filter(c *cli.Context) error {
	group := flagString(c, flagGroup)
	title := flagString(c, flagTitle)

	srv, err := load()
	if err != nil {
		return err
	}

	result, err := srv.Filter(group, title)
	if err != nil {
		return err
	}

	printRecords(result...)

	return nil
}

// delete removes an account with its history, or a whole group. Nothing asks
// for confirmation, -g takes a group with all the history in it.
func delete(c *cli.Context) error {
	group := flagString(c, flagGroup)
	title := flagString(c, flagTitle)
	if strings.TrimSpace(group) == "" && strings.TrimSpace(title) == "" {
		return errors.New("specify -g <group> or -t <title>")
	}

	srv, err := load()
	if err != nil {
		return err
	}

	if err := srv.Delete(group, title); err != nil {
		return err
	}

	return srv.Save()
}

func put(c *cli.Context) error {
	group := flagString(c, flagGroup)
	title := flagString(c, flagTitle)
	describe := flagString(c, flagDescribe)
	if strings.TrimSpace(title) == "" {
		return errors.New("title can't be empty")
	}

	srv, err := loadToWrite()
	if err != nil {
		return err
	}
	if err := unlockToken(srv); err != nil {
		return err
	}

	password, err := inputPassword()
	if err != nil {
		return err
	}

	record, err := srv.Put(group, title, password, describe)
	if err != nil {
		return err
	}

	// the record gets its sn on save
	if err := srv.Save(); err != nil {
		return err
	}

	fmt.Printf("saved: [%v] %v/%v\n", record.Sn, record.Group, record.Title)
	return nil
}

func get(c *cli.Context) error {
	target, err := getTarget(c)
	if err != nil {
		return err
	}

	srv, err := load()
	if err != nil {
		return err
	}
	if err := unlockToken(srv); err != nil {
		return err
	}

	var result *data.Record
	if target.bySn {
		result, err = srv.GetBySn(target.sn)
	} else {
		result, err = srv.Get(target.title)
	}
	if err != nil {
		return err
	}

	// shown with the password, so a sn that moved since the last list is seen here
	fmt.Printf("[%v] %v/%v\n", result.Sn, result.Group, result.Title)
	if flagBool(c, flagPrint) {
		fmt.Println(result.Password)
		return nil
	}

	if err := clipboard.WriteAll(result.Password); err != nil {
		return err
	}
	fmt.Println("your password has copied to clipboard")
	return nil
}

// target is how `get` finds an account.
type target struct {
	bySn  bool
	sn    int
	title string
}

// getTarget reads the arguments of `get`. A title may be all digits too, so a
// sn has to be asked for with -n, there is no guessing.
func getTarget(c *cli.Context) (*target, error) {
	bySn, byTitle := flagIsSet(c, flagSn), flagIsSet(c, flagTitle)
	if bySn == byTitle {
		return nil, errors.New("specify exactly one of -t <title> or -n <sn>")
	}
	if bySn {
		return &target{bySn: true, sn: flagInt(c, flagSn)}, nil
	}
	return &target{title: flagString(c, flagTitle)}, nil
}

func history(c *cli.Context) error {
	title := flagString(c, flagTitle)

	srv, err := load()
	if err != nil {
		return err
	}
	if err := unlockToken(srv); err != nil {
		return err
	}

	result, err := srv.History(title)
	if err != nil {
		return err
	}

	printRecordsV(result...)

	return nil
}

func resetKey(c *cli.Context) error {
	srv, err := load()
	if err != nil {
		return err
	}

	// the old token has to be right before anything is touched
	if err := unlockToken(srv); err != nil {
		return err
	}

	fmt.Println("please set your new key below")

	newPincode, err := inputPincodeTwice()
	if err != nil {
		return err
	}
	newToken, err := inputTokenTwice()
	if err != nil {
		return err
	}

	if err := srv.ResetKey(newPincode, newToken); err != nil {
		return err
	}

	return srv.Save()
}

// upgrade turns a v1 store into v2, see manager.SrvApi.Upgrade.
func upgrade(c *cli.Context) error {
	srv := manager.New()
	state, err := srv.State()
	if err != nil {
		return err
	}

	switch state {
	case manager.Missing:
		return errs.Uninited
	case manager.Blank:
		return errs.EmptyStore
	case manager.Unknown:
		return errs.InvalidFormat
	case manager.Current:
		fmt.Println("the store is already up to date")
		return nil
	}

	pincode, err := inputPincode()
	if err != nil {
		return err
	}
	token, err := inputToken()
	if err != nil {
		return err
	}

	backup, err := srv.Upgrade(pincode, token)
	if err != nil {
		return err
	}

	fmt.Printf("the store is upgraded, the v1 one is kept in %v\n", backup)
	// v1 derived its keys without a slow kdf, the v1 file, and its backup, is
	// easy to brute force. The factors found there open the v2 store as well.
	fmt.Println("the pincode and the token are the v1 ones, and v1 is easy to crack: run `mypass key` to change them")
	return nil
}

func genkey(c *cli.Context) error {
	length := flagInt(c, flagLen)
	print := flagBool(c, flagPrint)

	if length <= 0 {
		return errors.New("length must be positive")
	}
	if length > maxKeyLen {
		return errors.New("password too long, not suggested")
	}

	password, err := randomPassword(length)
	if err != nil {
		return err
	}

	if print {
		fmt.Println(password)
		return nil
	} else {
		if err := clipboard.WriteAll(password); err != nil {
			return err
		}
		fmt.Println("your password has copied to clipboard")
		return nil
	}
}

const (
	maxKeyLen = 32

	keyAlphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ,./;'`[]-=<>?:{}+_()*&^%$#@!"
)

// randomPassword gives a password of unpredictable characters of keyAlphabet.
func randomPassword(length int) (string, error) {
	size := big.NewInt(int64(len(keyAlphabet)))

	b := make([]byte, length)
	for i := range b {
		// rand.Int picks every character with the same chance
		index, err := rand.Int(rand.Reader, size)
		if err != nil {
			return "", err
		}
		b[i] = keyAlphabet[index.Int64()]
	}
	return string(b), nil
}

// load opens the store with the pincode typed by the user, and fails on a store
// without records. A store the command can't work with is refused before
// anything is asked for.
func load() (manager.SrvApi, error) {
	srv, err := loadStore(false)
	if err != nil {
		return nil, err
	}

	empty, err := srv.Empty()
	if err != nil {
		return nil, err
	}
	if empty {
		return nil, errs.EmptyStore
	}
	return srv, nil
}

// loadToWrite opens the store for `put`, the one command that writes a blank
// store. The pincode typed for a blank store becomes the pincode of the store,
// so it is asked for twice.
func loadToWrite() (manager.SrvApi, error) {
	return loadStore(true)
}

// loadStore opens the store with the pincode typed by the user. A blank file
// is refused unless fresh is set, then it opens as a store without records.
func loadStore(fresh bool) (manager.SrvApi, error) {
	srv := manager.New()
	state, err := srv.State()
	if err != nil {
		return nil, err
	}

	switch state {
	case manager.Missing:
		return nil, errs.Uninited
	case manager.Legacy:
		return nil, errs.LegacyStore
	case manager.Unknown:
		return nil, errs.InvalidFormat
	case manager.Blank:
		if !fresh {
			return nil, errs.EmptyStore
		}
	}

	var pincode []byte
	if state == manager.Blank {
		pincode, err = inputPincodeTwice()
	} else {
		pincode, err = inputPincode()
	}
	if err != nil {
		return nil, err
	}

	return srv, srv.Load(pincode)
}

// unlockToken verifies the token typed by the user. A fresh store has no token
// yet, the one typed becomes its token, so it is asked for twice.
func unlockToken(srv manager.SrvApi) error {
	var token []byte
	var err error
	if srv.TokenSet() {
		token, err = inputToken()
	} else {
		token, err = inputTokenTwice()
	}
	if err != nil {
		return err
	}

	return srv.UnlockToken(token)
}
