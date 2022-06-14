(require :asdf)

(asdf:load-system :alexandria)
(asdf:load-system :com.inuoe.jzon)

(defun read-data ()
  (com.inuoe.jzon:parse #p"./data.json"))

(defun data-to-channels (data)
  (loop with channels = (make-hash-table :test #'equal)
        for entry across data
        collecting (list (gethash "channel" entry) (gethash "members" entry))))

(defun transpose-channels-to-members (channels)
  (loop with members-table = (make-hash-table :test #'equal)
        for (channel-name members) in channels
        do (add-channel-to-members channel-name members members-table)
        finally (return members-table)))

(defun add-channel-to-members (channel-name members members-table)
  (loop for member across members
        do (setf (gethash member members-table) (cons channel-name (gethash member members-table)))))

(defun add-channels-co-locations-from-members (members)
  (loop for member-channels in (alexandria:hash-table-values members)
        with co-locations = (make-hash-table :test #'equal)
        do (add-channels-co-locations-from-member co-locations member-channels)
        finally (return co-locations)))

(defun add-channels-co-locations-from-member (co-locations channels)
  (loop for xs on (sort channels #'string-lessp)
        do (loop for y in (cdr xs)
                 do (let ((name (format nil "~a<>~a" (car xs) y)))
                       (setf (gethash name co-locations) (+ 1 (or (gethash name co-locations) 0)))))))

(defun filter-co-locations (channels co-locations p-threshold t-threshold)
  (loop with new-co-location = (make-hash-table :test #'equal)
        for (k . v) in (alexandria:hash-table-alist co-locations)
        do (let* ((from (co-location-from k))
                  (to (co-location-to k))
                  (from-member-count (length (cadr (assoc from channels :test #'string=))))
                  (to-member-count (length (cadr (assoc to channels :test #'string=))))
                  (count (min from-member-count to-member-count)))
             (if (and (> (/ v count) p-threshold) (> v t-threshold))
                  (setf (gethash k new-co-location) (/ v count))))
        finally (return new-co-location)))

(defun channels-to-kumu-elements-csv (channels)
  (loop for channel in channels
        collecting (format nil "~s,~d" (car channel) (length (cadr channel))) into rows
        finally (return (format nil "Label,Size~%~{~a~%~}" rows))))

(defun channels-co-locations-to-kumu-connections-csv (co-locations)
  (loop for (k . v) in (alexandria:hash-table-alist co-locations)
        collecting (let ((from (car (split-str k "<"))) (to (cadr (split-str k ">"))))
                     (format nil "~s,~s,~f" from to v))
          into rows
        finally (return (format nil "From,To,Size~%~{~a~%~}" rows))))

(defun co-location-from (key)
  (car (split-str key "<")))

(defun co-location-to (key)
  (cadr (split-str key ">")))

(defun write-file (name content)
  (with-open-file (stream name
                          :direction :output
                          :if-exists :overwrite
                          :if-does-not-exist :create)
    (format stream content)))

(defun test-data ()
  (com.inuoe.jzon:parse "[{ \"channel\": \"foo\", \"members\": [ \"bar\", \"baz\" ] }]"))

(defun channels ()
  (data-to-channels (read-data)))

(defun members ()
  (transpose-channels-to-members (channels)))

(defun colocs ()
  (add-channels-co-locations-from-members (members)))

(defun filter-colocs (co-locations threshold)
  (loop for (k . v) in (alexandria:hash-table-alist co-locations)
        with new-colos = (make-hash-table :test #'equal)
        do (if (> v threshold)
               (setf (gethash k new-colos) v))
        finally (return new-colos)))



; Credit: https://gist.github.com/siguremon/1174988/babcbdcbbfcb9f42df34f000f9326a26caa64be4
(defun split-str (string &optional (separator " "))
  (split-str-1 string separator))

(defun split-str-1 (string &optional (separator " ") (r nil))
  (let ((n (position separator string
		     :from-end t
		     :test #'(lambda (x y)
			       (find y x :test #'string=)))))
    (if n
	(split-str-1 (subseq string 0 n) separator (cons (subseq string (1+ n)) r))
      (cons string r))))
