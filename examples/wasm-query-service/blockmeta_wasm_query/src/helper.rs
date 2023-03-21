pub fn parse_month(key: &String) -> String {
    let index = key.len() - 2;
    return key[index..].to_string();
}

#[cfg(test)]
mod tests {
    use crate::helper::parse_month;

    #[test]
    fn parse_month_test() {
        let str = "month:first:201909".to_string();
        assert_eq!("09", parse_month(&str));
    }
}
