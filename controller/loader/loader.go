/*
	Used to load all controller packages at once. Add needed accordingly.
*/
package loader

import (
	_ "upendo/controller"
	_ "upendo/controller/appcontrollers"
)

func init() { /* nothing to do here */ }
