package zypper

type XMLResult []byte

func GetInstalledProductsXML() (XMLResult, error) {
	args := []string{"--disable-repositories", "--xmlout", "--non-interactive", "products", "-i"}
	return zypperRun(args, []int{zypperOK})
}
