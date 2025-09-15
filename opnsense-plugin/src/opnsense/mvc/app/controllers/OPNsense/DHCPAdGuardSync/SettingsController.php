<?php
namespace OPNsense\DHCPAdGuardSync;

use OPNsense\Base\ApiMutableModelControllerBase;
use OPNsense\Core\Config;
use OPNsense\DHCPAdGuardSync\DHCPAdGuardSync;

class SettingsController extends ApiMutableModelControllerBase
{
    protected static $internalModelName = 'dhcpadguardsync';
    protected static $internalModelClass = '\OPNsense\DHCPAdGuardSync\DHCPAdGuardSync';

    public function indexAction()
    {
        // Include form definitions
        $this->view->generalForm = $this->getForm("general");
        $this->view->settingsForm = $this->getForm("dialogSettings");

        // Pick the template to serve
        $this->view->pick('OPNsense/DHCPAdGuardSync/settings');
    }
}
