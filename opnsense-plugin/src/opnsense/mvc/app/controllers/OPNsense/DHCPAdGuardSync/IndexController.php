<?php
namespace OPNsense\DHCPAdGuardSync;

use OPNsense\Core\Config;

class IndexController extends \OPNsense\Base\IndexController
{
    public function indexAction()
    {
        // Include form definitions
        $this->view->generalForm = $this->getForm("general");

        // Pick the template to serve
        $this->view->pick('OPNsense/DHCPAdGuardSync/index');
    }
}
