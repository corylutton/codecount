<?php
class Mhistory extends CI_Model {

    /*
    *  Add a history record for sending the email
    *
    */
    function create($ruleid, $jsondata)
    {
        $storedproc = "
            /* Notifier History create SQL */
            INSERT INTO ALPHA.R50ALL.CI_HISTORY
            (RULEID, JSONDATA)
            VALUES(?, ?)";

        $stmt = $this->db->query(
            $storedproc,
            array(
                $ruleid,
                str_replace("'", "''", $jsondata)
            )
        );
        return $stmt;

	}
}
# Blank = 3, Comment = 5,  Code = 19, Total = 27
