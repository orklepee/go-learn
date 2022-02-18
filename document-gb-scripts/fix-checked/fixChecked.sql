BEGIN TRY
	  DECLARE @upcount AS INT;
    UPDATE p SET p.CheckDate = pn.CreatedOn, p.CheckedBy = pn.CreatedBy,
                 p.Checked=(CASE WHEN p.Checked=0 THEN -1 WHEN p.Checked=2 THEN 3 ELSE p.Checked END)
    FROM Property p
             INNER JOIN (
        SELECT
            PropertyID,
            CreatedOn,
            CreatedBy,
            ROW_NUMBER() OVER (PARTITION BY PropertyID ORDER BY CreatedOn DESC) as rn
        FROM PropertyNotes
        WHERE NoteType = 'Checked'
          AND CreatedBy IN (SELECT initials From Users WHERE NTLogon LIKE 'tylerre%')
    ) AS pn
                        ON p.PropertyID = pn.PropertyID
    WHERE pn.rn=1
    AND p.CheckDate IN ('1/1/1989', NULL);
	  SELECT @upcount = @@ROWCOUNT;
	  
    UPDATE Property SET CheckDate = GETDATE(), CheckedBy = 'blg', Checked=-1
	  WHERE CheckDate IN ('1/1/1989', NULL) AND LEFT(LandUse,  1) = 'W' AND LandUse NOT IN ('W09', 'WAB')
	  SELECT @@ROWCOUNT + @upcount;
END TRY
BEGIN CATCH
    THROW
END CATCH