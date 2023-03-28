pub struct Date {
    pub year: u64,
    pub month: u64,
}

impl Date {
    pub fn incr(&mut self) {
        self.month += 1;
        if self.month > 12 {
            self.month = 1;
            self.year += 1;
        }
    }

    pub fn key(&self) -> String {
        return format!("{:0>2}{:0>2}", self.year.to_string(), self.month.to_string())
    }

}

pub fn parse_year(key: &String) -> String {
    let parts = key.split(":");
    let last = parts.last().unwrap();
    return last[0..4].to_string();
}


pub fn parse_month(key: &String) -> String {
    let parts = key.split(":");
    let last = parts.last().unwrap();
    return last[4..].to_string();
}

// expected format 2023-02
pub fn parse_date(date: String) -> Result<Date, String> {

    let parts = date.split("-").collect::<Vec<&str>>();

    if parts.len() != 2 {
        return Err("expected date format yyyy-mm".to_string());
    }

    let year = parts[0].parse::<u64>().unwrap();
    let month = parts[1].parse::<u64>().unwrap();

    if (month < 1) || (month > 12) {
        return Err(format!("month {} needs to be between 1 and 12", month))
    }
    Ok(Date{
        year,
        month
    })
}

#[cfg(test)]
mod tests {
    use crate::helper::{Date, parse_date, parse_month, parse_year};

    #[test]
    fn parse_month_test() {
        let str = "month:first:201909".to_string();
        assert_eq!("09", parse_month(&str));
    }

    #[test]
    fn parse_year_test() {
        let str = "month:first:201909".to_string();
        assert_eq!("2019", parse_year(&str));
    }

    #[test]
    fn parse_date_ok_one() {
        let resp  = parse_date("2020-2".to_string());
        assert!(resp.is_ok());
        let out  = resp.unwrap();
        assert_eq!(2020, out.year);
        assert_eq!(2, out.month);
    }

    #[test]
    fn parse_date_ok_two() {
        let resp  = parse_date("2020-11".to_string());
        assert!(resp.is_ok());
        let out  = resp.unwrap();
        assert_eq!(2020, out.year);
        assert_eq!(11, out.month);
    }
    #[test]
    fn parse_date_ok_three() {
        let resp  = parse_date("2021-03".to_string());
        assert!(resp.is_ok());
        let out  = resp.unwrap();
        assert_eq!(2021, out.year);
        assert_eq!(3, out.month);
    }

    #[test]
    fn parse_date_err_one() {
        let resp  = parse_date("2020-22".to_string());
        assert!(resp.is_err());
        assert_eq!("month 22 needs to be between 1 and 12", resp.err().unwrap());
    }

    #[test]
    fn parse_date_err_two() {
        let resp  = parse_date("202022".to_string());
        assert!(resp.is_err());
        assert_eq!("expected date format yyyy-mm", resp.err().unwrap());
    }

    #[test]
    fn date_key_1() {
        let d = Date {
            year: 2022,
            month: 2,
        };
        assert_eq!("202202", d.key())
    }

    #[test]
    fn date_key_2() {
        let d = Date {
            year: 2022,
            month: 12,
        };
        assert_eq!("202212", d.key())
    }
}
