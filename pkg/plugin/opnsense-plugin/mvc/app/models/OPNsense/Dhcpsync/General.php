<?php

namespace OPNsense\Dhcpsync;

use OPNsense\Base\BaseModel;

class General extends BaseModel
{
    /**
     * Check if the module is enabled
     * @return bool
     */
    public function isEnabled()
    {
        if (isset($this->enabled)) {
            return (string)$this->enabled === '1';
        }
        return false;
    }
}
